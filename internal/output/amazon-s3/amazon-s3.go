package s3

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//S3 struct
type S3 struct {
	config       Config
	Transactions <-chan transaction.Transaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
}

//Config struct
type Config struct {
	FilePath                      string `mapstructure:"filepath" validate:"required"`
	S3Region                      string `mapstructure:"s3_region" validate:"required"`
	S3Bucket                      string `mapstructure:"s3_bucket" validate:"required"`
	AccessKeyId                   string `mapstructure:"access_key_id"`
	SecretAccessKey               string `mapstructure:"secret_access_key"`
	SessionToken                  string `mapstructure:"session_token"`
	CannedAcl                     string `mapstructure:"canned_acl" validate:"oneof=private public-read public-read-write authenticated-read aws-exec-read bucket-owner-read bucket-owner-full-control log-delivery-write"`
	Encoding                      string `mapstructure:"encoding" validate:"oneof=none gzip"`
	ServerSideEncryptionAlgorithm string `mapstructure:"server_side_encryption_algorithm" validate:"oneof=AES256 aws:kms"`
	StorageClass                  string `mapstructure:"storage_class" validate:"oneof=STANDARD REDUCED_REDUNDANCY STANDARD_IA"`

	filepath config.Selector
}

//DefaultConfig func return the default configurations
func DefaultConfig() *Config {
	return &Config{
		CannedAcl:                     "private",
		Encoding:                      "none",
		ServerSideEncryptionAlgorithm: "AES256",
		StorageClass:                  "STANDARD",
	}
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &S3{}
}

//SetTransactionChan set Transaction chan that this plugin will use to receive transactions
func (s *S3) SetTransactionChan(t <-chan transaction.Transaction) {
	s.Transactions = t
}

//Init func Initialize the S3 output plugin
func (s *S3) Init(config config.Config, logger zap.SugaredLogger) error {

	var err error

	s.config = *DefaultConfig()
	err = config.Populate(&s.config)
	if err != nil {
		return err
	}

	s.config.filepath, err = config.NewSelector(s.config.FilePath)
	if err != nil {
		return err
	}

	s.stopChan = make(chan struct{})
	s.logger = logger

	return nil

}

//writeOnS3 func takes the transaction and session
//that to be written on the amazon S3
func (s *S3) writeOnS3(svc *s3.S3, txn transaction.Transaction) {
	defer s.wg.Done()

	var buffer []byte
	var err error

	switch Payload := txn.Payload.(type) {
	case payload.Bytes:
		buffer = Payload
	case payload.Stream:
		buffer, err = ioutil.ReadAll(Payload)
		if err != nil {
			txn.ResponseChan <- response.Error(err)
			return
		}
	}

	filePath, err := s.config.filepath.Evaluate(txn.Data)
	if err != nil {
		txn.ResponseChan <- response.Error(err)
		return
	}

	size := int64(len(buffer))
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s.config.S3Bucket),
		Key:                  aws.String(filePath),
		ACL:                  aws.String(s.config.CannedAcl),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ContentEncoding:      aws.String(s.config.Encoding),
		ServerSideEncryption: aws.String(s.config.ServerSideEncryptionAlgorithm),
		StorageClass:         aws.String(s.config.StorageClass),
	})

	if err != nil {
		txn.ResponseChan <- response.Error(err)
		return
	}

	// send response
	txn.ResponseChan <- response.Ack()
}

// Start the plugin and be ready for taking transactions
func (s *S3) Start() error {

	staticCreds := credentials.NewStaticCredentials(s.config.AccessKeyId, s.config.SecretAccessKey, s.config.SessionToken)
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
	cfg := aws.NewConfig().WithRegion(s.config.S3Region).WithCredentials(creds)

	sess, err := session.NewSession(cfg)
	if err != nil {
		return err
	}

	svc := s3.New(sess, cfg)

	//Test if the given credentials are valid or not by getting the bucket logging
	bucketName := s.config.S3Bucket
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
			case txn := <-s.Transactions:
				s.wg.Add(1)
				go s.writeOnS3(svc, txn)
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
