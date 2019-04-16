package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

type Writer struct {
	bytes.Buffer
	eof      error
	eofTotal int
}

func (r *Writer) Close() error {
	r.eofTotal = r.Len()
	r.eof = errors.New("EOF")
	return nil
}

type Clone struct {
	source *Writer
	i      int
}

func NewClone(source *Writer) *Clone {
	return &Clone{
		source: source,
		i:      0,
	}
}

func (c *Clone) Read(p []byte) (read int, eof error) {
	upperlimit := c.i + len(p)
	if upperlimit > c.source.Len() {
		upperlimit = c.source.Len()
		// check if EOF
		if upperlimit >= c.source.eofTotal {
			eof = c.source.eof
		}
	}

	copy(p, c.source.Bytes()[c.i:upperlimit])

	read = upperlimit - c.i
	c.i = upperlimit

	return read, eof
}

func main() {

	r := Writer{}

	go func() {
		pulser := time.Tick(1000 * time.Millisecond)
		breaker := time.After(10000 * time.Millisecond)
		i := 0
	outerloop:
		for {
			select {
			case <-pulser:
				_, _ = fmt.Fprint(&r, "xxxxxxxxx-xxxxxxxxx-xxxxxxxxx"+strconv.Itoa(i))
				i++
			case <-breaker:
				log.Println("Done writing...")
				break outerloop

			}
		}
		r.Close()
	}()

	c1 := NewClone(&r)
	c2 := NewClone(&r)

	pulser := time.Tick(300 * time.Millisecond)
	for {
		select {
		case <-pulser:
			buff := make([]byte, 5)
			n, eof := c1.Read(buff)
			if eof != nil {
				goto breakloop
			}
			if n > 0 {
				log.Println("C1", n, eof)
				log.Println(string(buff[:n]))
			}
			n, eof = c2.Read(buff)
			if eof != nil {
				goto breakloop
			}
			if n > 0 {
				log.Println("C2", n, eof)
				log.Println(string(buff[:n]))
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
