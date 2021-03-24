package main

import (
	"flag"
	"os"
	"testing"

	feature "github.com/ONSdigital/dp-observation-importer/features/steps"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

func InitializeScenario(ctx *godog.ScenarioContext) {
	importerFeature := feature.NewObservationImporterFeature()

	ctx.BeforeScenario(func(*godog.Scenario) {
		importerFeature.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		importerFeature.Close()
	})

	importerFeature.RegisterSteps(ctx)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})
	ctx.AfterSuite(func() {
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
		}

		status := godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  InitializeScenario,
			TestSuiteInitializer: InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}

	} else {
		t.Skip("component flag required to run component tests")
	}
}
