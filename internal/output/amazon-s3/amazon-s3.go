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
	"sync"
	"time"
)

//S3 struct
type S3 struct {
	Settings     map[string]string
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
	s.Settings = make(map[string]string)
	requiredConfig := []string{"filepath", "s3_region", "s3_bucket"}
	for _, rq := range requiredConfig {
		value, err := config.Get(rq, nil)
		if err != nil {
			return err
		}
		if value.String() == "" {
			err = errors.New(rq + " Cannot be empty")
			return err
		}
		s.Settings[rq] = value.String()
	}
	optionalConfig := []string{"access_key_id", "secret_access_key", "session_token", "canned_acl", "encoding", "server_side_encryption_algorithm", "storage_class"}
	s.Settings["canned_acl"] = "private"
	s.Settings["encoding"] = "none"
	s.Settings["server_side_encryption_algorithm"] = "AES256"
	s.Settings["storage_class"] = "STANDARD"
	for _, op := range optionalConfig {
		value, err := config.Get(op, nil)
		if err != nil {
			continue
		}
		if value.String() == "" {
			err = errors.New(op + " Cannot be empty")
			return err
		}
		s.Settings[op] = value.String()
	}
	s.Transactions = make(chan types.Transaction)
	s.stopChan = make(chan struct{})
	s.logger = logger
	return nil
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
			Bucket:               aws.String(s.Settings["s3_bucket"]),
			Key:                  aws.String(s.Settings["filepath"]),
			ACL:                  aws.String(s.Settings["canned_acl"]),
			Body:                 bytes.NewReader(buffer),
			ContentLength:        aws.Int64(size),
			ContentType:          aws.String(http.DetectContentType(buffer)),
			ContentDisposition:   aws.String("attachment"),
			ContentEncoding:      aws.String(s.Settings["encoding"]),
			ServerSideEncryption: aws.String(s.Settings["server_side_encryption_algorithm"]),
			StorageClass:         aws.String(s.Settings["storage_class"]),
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

	staticCreds := credentials.NewStaticCredentials(s.Settings["access_key_id"], s.Settings["secret_access_key"], s.Settings["session_token"])
	val, _ := staticCreds.Get()
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.StaticProvider{
				Value: val,
			},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
		})
	_, err := creds.Get()
	if err != nil {
		return err
	}
	cfg := aws.NewConfig().WithRegion(s.Settings["s3_region"]).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	//Test if the given credentials are valid or not by getting the bucket logging
	bucketName := s.Settings["s3_bucket"]
	tst := s3.GetBucketLoggingInput{Bucket: &bucketName}
	_, err = svc.GetBucketLogging(&tst)
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
