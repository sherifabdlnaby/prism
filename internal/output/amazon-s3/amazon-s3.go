package s3

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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
	FilePath        string
	Region          string
	Bucket          string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	TypeCheck       bool
	Transactions    chan types.Transaction
	stopChan        chan struct{}
	logger          zap.Logger
	wg              sync.WaitGroup
}

//TransactionChan just return the channel of the transactions
func (s *S3) TransactionChan() chan<- types.Transaction {
	return s.Transactions
}

//Init func Initialize the S3 output plugin
func (s *S3) Init(config types.Config, logger zap.Logger) error {
	var err error
	value, dummyErr := config.Get("filepath", nil)
	err = dummyErr
	if err == nil {
		s.FilePath = value.String()
		if s.FilePath == "" && err == nil {
			err = errors.New("FilePath cannot be empty")
		}
		path.Clean(s.FilePath)
	}
	if err == nil {
		value, dummyErr := config.Get("s3_region", nil)
		err = dummyErr
		s.Region = value.String()
		if s.Region == "" && err == nil {
			err = errors.New("S3_Region cannot be empty")
		}
	}
	if err == nil {
		value, dummyErr := config.Get("s3_bucket", nil)
		err = dummyErr
		s.Bucket = value.String()
		if s.Bucket == "" && err == nil {
			err = errors.New("S3_Bucket cannot be empty")
		}
	}
	if err == nil {
		value, dummyErr := config.Get("access_key_id", nil)
		if dummyErr == nil {
			s.AccessKeyId = value.String()
			if s.AccessKeyId == "" {
				err = errors.New("S3 AWS acess key Id cannot be empty")
			}
		}
	}
	if err == nil {
		value, dummyErr := config.Get("secret_access_key", nil)
		if dummyErr == nil {
			s.SecretAccessKey = value.String()
			if s.SecretAccessKey == "" {
				err = errors.New("S3 AWS secret acess key cannot be empty")
			}
		}
	}
	if err == nil {
		value, dummyErr := config.Get("session_token", nil)
		if dummyErr == nil {
			s.Token = value.String()
			if s.Token == "" {
				err = errors.New("S3 Session Token cannot be empty")
			}
		}
	}
	s.Transactions = make(chan types.Transaction)
	s.stopChan = make(chan struct{})
	s.logger = logger
	return err
}

//writeOnS3 func takes the transaction and session
//that to be written on the amazon S3
func (s *S3) writeOnS3(svc *s3.S3, transaction types.Transaction) {
	defer s.wg.Done()
	ack := true
	buffer, err := ioutil.ReadAll(transaction)
	size := int64(len(buffer))
	if err == nil {
		_, err = svc.PutObject(&s3.PutObjectInput{
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

	staticCreds := credentials.NewStaticCredentials(s.AccessKeyId, s.SecretAccessKey, "")
	val, _ := staticCreds.Get()
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&credentials.StaticProvider{
				Value: val,
			},
		})
	_, err := creds.Get()
	if err != nil {
		return err
	}
	cfg := aws.NewConfig().WithRegion(s.Region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.stopChan:
				return
			case transaction, _ := <-s.Transactions:
				s.wg.Add(1)
				go s.writeOnS3(svc, transaction)
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
