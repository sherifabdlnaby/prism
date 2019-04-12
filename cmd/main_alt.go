package main

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"gopkg.in/yaml.v2"
	"log"
)

func main() {

	appConfig := config.PipelinesConfig{}

	err := config.Load("pipeline.yaml", &appConfig, true)

	fmt.Printf("--- t:\n%v\n\n", appConfig)

	d, err := yaml.Marshal(&appConfig)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("--- t dump:\n%s\n\n", string(d))

}
