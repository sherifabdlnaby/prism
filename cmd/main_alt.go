package main

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"gopkg.in/yaml.v2"
	"log"
)

func main() {

	appConfig := config.PipelinesConfig{}

	err := config.Load("pipeline.yaml", &appConfig, true)

	fmt.Printf("--- t:\n%v\n\n", appConfig)

	d, err := yaml.Marshal(&appConfig)

	pipelineX := pipeline.NewPipeline(appConfig.Pipelines["profile_pic_pipeline"])

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("--- t dump:\n%s\n\n", string(d))
	fmt.Printf("--- t dump:\n%v\n\n", pipelineX)

}
