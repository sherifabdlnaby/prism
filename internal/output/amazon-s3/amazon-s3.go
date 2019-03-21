package amazon_s3

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"path"
	"sync"
	"time"
)

//S3 struct
type S3 struct {
	FilePath     string
	Region       string
	Bucket       string
	TypeCheck    bool
	Transactions chan types.Transaction
	stopChan     chan struct{}
	logger       zap.Logger
	wg           sync.WaitGroup
}

//TransactionChan just return the channel of the transactions
func (s *S3) TransactionChan() chan<- types.Transaction {
	return s.Transactions
}

//Init func Initialize the S3 output plugin
func (s *S3) Init(config types.Config, logger zap.Logger) error {
	if s.FilePath, s.TypeCheck = config["filepath"].(string); !s.TypeCheck {
		return errors.New("FilePath must be a string")
	}
	if s.FilePath == "" {
		return errors.New("FilePath cannot be empty")
	}
	path.Clean(s.FilePath)
	if s.Region, s.TypeCheck = config["s3_region"].(string); !s.TypeCheck {
		return errors.New("S3_Region must be a string")
	}
	if s.Region == "" {
		return errors.New("S3_Region cannot be empty")
	}
	if s.Bucket, s.TypeCheck = config["s3_bucket"].(string); !s.TypeCheck {
		return errors.New("S3_Bucket must be a string")
	}
	if s.Bucket == "" {
		return errors.New("S3_Bucket cannot be empty")
	}
	s.Transactions = make(chan types.Transaction)
	s.stopChan = make(chan struct{})
	s.logger = logger
	return nil
}

//writeOnS3 func takes the transaction and session
//that to be written on the amazon S3
func (s *S3) writeOnS3(session *session.Session, transaction types.Transaction) {
	defer s.wg.Done()
	ack := true
	buffer, err := ioutil.ReadAll(transaction)
	size := int64(len(buffer))
	if err == nil {
		_, err = s3.New(session).PutObject(&s3.PutObjectInput{
			Bucket:               aws.String(s.Bucket),
			Key:                  aws.String(s.FilePath),
			ACL:                  aws.String("private"),
			Body:                 bytes.NewReader(buffer),
			ContentLength:        aws.Int64(size),
			ContentType:          aws.String(http.DetectContentType(buffer)),
			ContentDisposition:   aws.String("attachment"),
			ServerSideEncryption: aws.String("AES256"),
		})
	}
	if err != nil {
		ack = false
	}
	// send response
	transaction.ResponseChan <- types.Response{
		Error: err,
		Ack:   ack,
	}
	return
}

// Start the plugin and be ready for taking transactions
func (s *S3) Start() error {
	mySession, err := session.NewSession(&aws.Config{Region: aws.String(s.Region)})
	if err != nil {
		return err
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.stopChan:
				return
			case transaction, _ := <-s.Transactions:
				s.wg.Add(1)
				go s.writeOnS3(mySession, transaction)
			}
		}
	}()

	return nil
}

//Close func Send a close signal to stop chan
// to stop taking transactions and Close everything safely
func (s *S3) Close(time.Duration) error {
	s.stopChan <- struct{}{}
	s.wg.Wait()
	close(s.Transactions)
	close(s.stopChan)
	return nil
}
