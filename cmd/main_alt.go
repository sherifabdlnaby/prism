package main

import (
	"bytes"
	"fmt"
	"io"
)

// buffer is just here to make bytes.Buffer an io.ReadWriteCloser.
// Read about embedding to see how this works.
type buffer struct {
	bytes.Buffer
}

// Add a Close method to our buffer so that we satisfy io.ReadWriteCloser.
func (b *buffer) Close() error {
	b.Buffer.Reset()
	return nil
}

func main() {

	// Address the OP question.
	var rwc io.ReadWriteCloser

	// Make the io.ReadWriteCloser actually do something.
	rwc = &buffer{}

	// Write some bytes to the buffer. We could also do this by:
	//  n, err := rwc.Write([]byte("hello")
	// where n is the number of bytes successfully written and
	// err is any error that happened during the write (fmt.Fprint
	// will give these too if you want).
	fmt.Fprint(rwc, "11111-")
	fmt.Fprint(rwc, "22222-")
	// Close the buffer. This could have been defer'd above.
	rwc.Close()
	fmt.Fprint(rwc, "33333----")

	// This is a byte slice we will fill with the read.
	// It is longer that the contents of the buffer.
	// Try making it shorter to see what happens - will
	// the following loop be correct? If not, how do you
	// fix it?
	buf := make([]byte, 10)

	// Here we do two reads of the ReadWriteCloser
	// and show all the returned values
	// and the byte slice that is being filled.
	for {
		n, err := rwc.Read(buf)
		fmt.Printf("read %d bytes %q got a %v error (total buffer is %v)\n", n, buf[:n], err, buf)
		if err != nil {
			break
		}
	}

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
