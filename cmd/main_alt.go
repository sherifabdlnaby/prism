package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/sherifabdlnaby/prism/pkg/mirror"
)

func main() {

	r := mirror.Writer{}

	go func() {
		pulser := time.Tick(1000 * time.Millisecond)
		breaker := time.After(10000 * time.Millisecond)
		i := 0
	outerloop:
		for {
			select {
			case <-pulser:
				ns, _ := fmt.Fprint(&r, "xxxxxxxxx-xxxxxxxxx-xxxxxxxxx"+strconv.Itoa(i))
				log.Println(ns)
				i++
			case <-breaker:
				log.Println("Done writing...")
				break outerloop

			}
		}
		r.Close()
	}()

	time.Sleep(1000 * time.Millisecond)

	rdr := r.NewReader()
	mrr := mirror.NewReader(rdr)
	c1 := mrr.Clone()
	c2 := mrr.Clone()

	pulser := time.Tick(1000 * time.Millisecond)
	for {
		select {
		case <-pulser:
			buff := make([]byte, 5000)
			n, eof := c1.Read(buff)
			if n > 0 {
				log.Println("C1", n, eof)
				log.Println(string(buff[:n]))
			}
			n, eof = c2.Read(buff)
			if n > 0 {
				log.Println("C2", n, eof)
				log.Println(string(buff[:n]))
			}
			if eof != nil {
				goto breakloop
			}
		}

	}

breakloop:
	return

	//appConfig := config.PipelinesConfig{}
	//
	//err := config.Load("pipeline.yaml", &appConfig, true)
	//
	//fmt.Printf("--- t:\n%v\n\n", appConfig)
	//
	//d, err := yaml.Marshal(&appConfig)
	//
	//pipelineX := pipeline.NewPipeline(appConfig.Pipelines["profile_pic_pipeline"])
	//
	//if err != nil {
	//	log.Fatalf("error: %v", err)
	//}
	//
	//fmt.Printf("--- t dump:\n%s\n\n", string(d))
	//fmt.Printf("--- t dump:\n%v\n\n", pipelineX)

}
