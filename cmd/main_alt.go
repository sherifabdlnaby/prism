package main

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/stretchr/objx"
	"gopkg.in/yaml.v2"
	"log"
)

func main() {

	appConfig := config.InputsConfig{}
	//appConfig := make(map[string]interface{})

	err := config.Load("input_plugins.yaml", &appConfig, true)
	fmt.Println(appConfig.Inputs)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	intr := map[string]interface{}(appConfig.Inputs["http_server"].Config)
	lol := objx.New(intr)

	xx := lol.Get("host").String()
	xx = lol.Get("urls./cover_picture.pipeline").String()

	intr2 := map[string]interface{}(appConfig.Inputs["http_server2"].Config)
	lol2 := objx.New(intr2)

	xx2 := lol2.Get("host").String()

	fmt.Println(xx)
	fmt.Println(xx2)

	fmt.Printf("--- t:\n%v\n\n", appConfig)

	d, err := yaml.Marshal(&appConfig)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("--- t dump:\n%s\n\n", string(d))

}
