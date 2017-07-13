package kafka

import (
	"os"

	"fmt"
	"github.com/ONSdigital/dp-publish-pipeline/utils"
	"github.com/ONSdigital/go-ns/log"
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"time"
)

var tick = time.Millisecond * 4000

// NOTE: to be replaced by go-ns kafka library.

type ConsumerGroup struct {
	consumer *cluster.Consumer
	incoming chan Message
	closer   chan bool
	errors   chan error
}

func (consumer ConsumerGroup) Incoming() chan Message {
	return consumer.incoming
}

func (consumer ConsumerGroup) Closer() chan bool {
	return consumer.closer
}

type Message interface {
	GetData() []byte
	Commit()
}

type SaramaMessage struct {
	message  *sarama.ConsumerMessage
	consumer *cluster.Consumer
}

func (M SaramaMessage) GetData() []byte {
	return M.message.Value
}

func (M SaramaMessage) Commit() {
	M.consumer.MarkOffset(M.message, "metadata")
}

func SetMaxMessageSize(maxSize int32) {
	sarama.MaxRequestSize = maxSize
	sarama.MaxResponseSize = maxSize
}

func NewConsumerGroup(topic string, group string) (*ConsumerGroup, error) {
	config := cluster.NewConfig()
	config.Group.Return.Notifications = true
	config.Consumer.Return.Errors = true
	config.Consumer.MaxWaitTime = 50 * time.Millisecond
	brokers := []string{utils.GetEnvironmentVariable("KAFKA_ADDR", "localhost:9092")}

	consumer, err := cluster.NewConsumer(brokers, group, []string{topic}, config)
	if err != nil {
		return nil, fmt.Errorf("Bad NewConsumer of %q: %s", topic, err)
	}

	cg := ConsumerGroup{
		consumer: consumer,
		incoming: make(chan Message),
		closer:   make(chan bool),
		errors:   make(chan error),
	}
	signals := make(chan os.Signal, 1)
	//signal.Notify(signals, os.)

	go func() {
		defer cg.consumer.Close()
		log.Info(fmt.Sprintf("Started kafka consumer of topic %q group %q", topic, group), nil)
		for {
			select {
			case err := <-cg.consumer.Errors():
				log.Error(err, nil)
				cg.errors <- err
			default:
				select {
				case msg := <-cg.consumer.Messages():
					cg.incoming <- SaramaMessage{msg, cg.consumer}
				case n, more := <-cg.consumer.Notifications():
					if more {
						log.Trace("Rebalancing group", log.Data{"topic": topic, "group": group, "partitions": n.Current[topic]})
					}
				case <-time.After(tick):
					cg.consumer.CommitOffsets()
				case <-signals:
					log.Info(fmt.Sprintf("Quitting kafka consumer of topic %q group %q", topic, group), nil)
					return
				case <-cg.closer:
					log.Info(fmt.Sprintf("Closing kafka consumer of topic %q group %q", topic, group), nil)
					return
				}
			}
		}
	}()
	return &cg, nil
}
