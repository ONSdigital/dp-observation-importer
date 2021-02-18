package feature

import (
	"github.com/cucumber/godog"
)

type ImporterFeature struct {
}

func NewObservationImporterFeature(url string) *ImporterFeature {
	return &ImporterFeature{}
}

func (f *ImporterFeature) RegisterSteps(context *godog.ScenarioContext) {

}

func (f *ImporterFeature) Close() {

}

func (f *ImporterFeature) Reset() {

}
