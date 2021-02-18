package main

import (
	"flag"
	feature "github.com/ONSdigital/dp-observation-importer/features/steps"
	"os"
	"testing"

	featuretest "github.com/armakuni/dp-go-featuretest"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type FeatureTest struct {
	Mongo *featuretest.MongoCapability
}

func (f *FeatureTest) InitializeScenario(ctx *godog.ScenarioContext) {
	authorizationFeature := featuretest.NewAuthorizationFeature()
	importerFeature := feature.NewObservationImporterFeature(authorizationFeature.FakeAuthService.ResolveURL(""))

	ctx.BeforeScenario(func(*godog.Scenario) {
		importerFeature.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		importerFeature.Close()
	})

	importerFeature.RegisterSteps(ctx)
	authorizationFeature.RegisterSteps(ctx)
}

func (f *FeatureTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})
	ctx.AfterSuite(func() {
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
		}

		f := &FeatureTest{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		os.Exit(status)
	} else {
		t.Skip("component flag required to run component tests")
	}
}
