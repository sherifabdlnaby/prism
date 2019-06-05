package s3

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//S3 struct
type S3 struct {
	Settings     map[string]config.Value
	Values       map[string]string
	TypeCheck    bool
	Transactions <-chan transaction.Transaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
}

// NewComponent Return a new Component
func NewComponent() processor.Component {
	return &S3{}
}

//TransactionChan just return the channel of the transactions
func (s *S3) TransactionChan(t <-chan transaction.Transaction) {
	s.Transactions = t
}

//Init func Initialize the S3 output plugin
func (s *S3) Init(Config config.Config, logger zap.SugaredLogger) error {

	s.Settings = make(map[string]config.Value)
	s.Values = make(map[string]string)

	requiredConfig := []string{"filepath", "s3_region", "s3_bucket"}

	for _, rq := range requiredConfig {
		value, err := Config.Get(rq, nil)
		if err != nil {
			return err
		}
		if value.Get().String() == "" {
			err = errors.New(rq + " Cannot be empty")
			return err
		}
		s.Settings[rq] = value
		s.Values[rq] = value.Get().String()
	}

	optionalConfig := []string{"access_key_id", "secret_access_key", "session_token", "canned_acl", "encoding", "server_side_encryption_algorithm", "storage_class"}
	s.Settings["canned_acl"] = Config.NewValue("private")
	s.Values["canned_acl"] = "private"
	s.Settings["encoding"] = Config.NewValue("none")
	s.Values["encoding"] = "node"
	s.Settings["server_side_encryption_algorithm"] = Config.NewValue("AES256")
	s.Values["server_side_encryption_algorithm"] = "AES256"
	s.Settings["storage_class"] = Config.NewValue("STANDARD")
	s.Values["storage_class"] = "STANDARD"
	for _, op := range optionalConfig {
		value, err := Config.Get(op, nil)
		if err != nil {
			continue
		}
		if value.Get().String() == "" {
			err = errors.New(op + " Cannot be empty")
			return err
		}
		s.Settings[op] = value
		s.Values[op] = value.Get().String()
	}
	s.stopChan = make(chan struct{})
	s.logger = logger

	return nil
}

//writeOnS3 func takes the transaction and session
//that to be written on the amazon S3
func (s *S3) writeOnS3(svc *s3.S3, txn transaction.Transaction) {

	defer s.wg.Done()
	ack := true

	buffer, err := ioutil.ReadAll(txn)
	size := int64(len(buffer))

	FilePathValue := s.Settings["filepath"]
	filePathV, evaluateErr := (&FilePathValue).Evaluate(txn.Data)
	if evaluateErr != nil {
		txn.ResponseChan <- response.Error(evaluateErr)
		return
	}
	filePath := filePathV.String()

	if err == nil {
		_, err = svc.PutObject(&s3.PutObjectInput{
			Bucket:               aws.String(s.Values["s3_bucket"]),
			Key:                  aws.String(filePath),
			ACL:                  aws.String(s.Values["canned_acl"]),
			Body:                 bytes.NewReader(buffer),
			ContentLength:        aws.Int64(size),
			ContentType:          aws.String(http.DetectContentType(buffer)),
			ContentDisposition:   aws.String("attachment"),
			ContentEncoding:      aws.String(s.Values["encoding"]),
			ServerSideEncryption: aws.String(s.Values["server_side_encryption_algorithm"]),
			StorageClass:         aws.String(s.Values["storage_class"]),
		})
	}
	if err != nil {
		ack = false
	}
	// send response
	txn.ResponseChan <- response.Response{
		Error: err,
		Ack:   ack,
	}
}

// Start the plugin and be ready for taking transactions
func (s *S3) Start() error {

	staticCreds := credentials.NewStaticCredentials(s.Values["access_key_id"], s.Values["secret_access_key"], s.Values["session_token"])
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
	cfg := aws.NewConfig().WithRegion(s.Values["s3_region"]).WithCredentials(creds)

	sess, err := session.NewSession(cfg)
	if err != nil {
		return err
	}

	svc := s3.New(sess, cfg)

	//Test if the given credentials are valid or not by getting the bucket logging
	bucketName := s.Values["s3_bucket"]
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
			case transaction := <-s.Transactions:
				s.wg.Add(1)
				go s.writeOnS3(svc, transaction)
			}
		}
	}()

	return nil
}

//Close func Send a close signal to stop chan
// to stop taking transactions and Close everything safely
func (s *S3) Close() error {
	s.stopChan <- struct{}{}
	s.wg.Wait()
	close(s.stopChan)
	return nil
}
