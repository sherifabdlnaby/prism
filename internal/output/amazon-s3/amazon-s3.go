package s3

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sync"
)

//S3 struct
type S3 struct {
	Settings     map[string]config.Value
	TypeCheck    bool
	Transactions chan transaction.Transaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &S3{}
}

//TransactionChan just return the channel of the transactions
func (s *S3) TransactionChan() chan<- transaction.Transaction {
	return s.Transactions
}

//Init func Initialize the S3 output plugin
func (s *S3) Init(Config config.Config, logger zap.SugaredLogger) error {

	s.Settings = make(map[string]config.Value)

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
	}

	optionalConfig := []string{"access_key_id", "secret_access_key", "session_token", "canned_acl", "encoding", "server_side_encryption_algorithm", "storage_class"}
	/*s.Settings["canned_acl"] = "private"
	s.Settings["encoding"] = "none"
	s.Settings["server_side_encryption_algorithm"] = "AES256"
	s.Settings["storage_class"] = "STANDARD"*/
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
	}
	s.Transactions = make(chan transaction.Transaction)
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

	if err == nil {
		_, err = svc.PutObject(&s3.PutObjectInput{
			Bucket:               aws.String(s.Settings["s3_bucket"].Get().String()),
			Key:                  aws.String(s.Settings["filepath"].Get().String()),
			ACL:                  aws.String(s.Settings["canned_acl"].Get().String()),
			Body:                 bytes.NewReader(buffer),
			ContentLength:        aws.Int64(size),
			ContentType:          aws.String(http.DetectContentType(buffer)),
			ContentDisposition:   aws.String("attachment"),
			ContentEncoding:      aws.String(s.Settings["encoding"].Get().String()),
			ServerSideEncryption: aws.String(s.Settings["server_side_encryption_algorithm"].Get().String()),
			StorageClass:         aws.String(s.Settings["storage_class"].Get().String()),
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
	return
}

// Start the plugin and be ready for taking transactions
func (s *S3) Start() error {

	staticCreds := credentials.NewStaticCredentials(s.Settings["access_key_id"].Get().String(), s.Settings["secret_access_key"].Get().String(), s.Settings["session_token"].Get().String())
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
	cfg := aws.NewConfig().WithRegion(s.Settings["s3_region"].Get().String()).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	//Test if the given credentials are valid or not by getting the bucket logging
	bucketName := s.Settings["s3_bucket"].Get().String()
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
func (s *S3) Close() error {
	s.stopChan <- struct{}{}
	s.wg.Wait()
	close(s.Transactions)
	close(s.stopChan)
	return nil
}
