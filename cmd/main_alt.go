package main

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"gopkg.in/yaml.v2"
	"log"
)

func main() {

	appConfig := config.InputsConfig{}
	err := config.Load("input_plugins.yaml", &appConfig, true)

	fmt.Println(appConfig.Inputs)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	test := types.NewConfig(appConfig.Inputs["http_server"].Config)

	data := map[string]interface{}{
		"port":   12,
		"length": 1,
		"width":  2,
		"test":   2,
	}

	data2 := map[string]interface{}{
		"port":   3711,
		"length": 2138,
		"width":  1908,
		"test":   912,
		"nested": data,
	}

	val, err := test.Get("port", data2)

	if err != nil {
		panic(err)
	}

	x := val.String()
	y := val.Int()
	z := val.Float32()

	fmt.Println(x, y, z)

	fmt.Printf("--- t:\n%v\n\n", appConfig)

	d, err := yaml.Marshal(&appConfig)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("--- t dump:\n%s\n\n", string(d))

}
