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
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//S3 struct
type S3 struct {
	config   config
	jobChan  <-chan job.Job
	stopChan chan struct{}
	logger   zap.SugaredLogger
	wg       sync.WaitGroup
}

// NewComponent Return a new Base
func NewComponent() component.Base {
	return &S3{}
}

//SetJobChan set Job chan that this plugin will use to receive jobs
func (s *S3) SetJobChan(t <-chan job.Job) {
	s.jobChan = t
}

//Init func Initialize the S3 output plugin
func (s *S3) Init(config cfg.Config, logger zap.SugaredLogger) error {

	s.config = *defaultConfig()
	Error := config.Populate(&s.config)
	if Error != nil {
		return Error
	}

	var err error
	s.config.filepath, err = config.NewSelector(s.config.FilePath)
	if err != nil {
		return err
	}

	s.stopChan = make(chan struct{})
	s.logger = logger

	return nil

}

// Start the plugin and be ready for taking jobs
func (s *S3) Start() error {

	creds, err := getCredentials(s)
	if err != nil {
		return err
	}

	client, err := getClient(s.config.S3Region, creds)
	if err != nil {
		return err
	}

	//Test if the given credentials are valid or not by getting the bucket logging
	err = pingBucket(s.config.S3Bucket, client)
	if err != nil {
		return err
	}

	go func() {
		for Job := range s.jobChan {
			s.wg.Add(1)
			go s.writeOnS3(client, Job)
		}
	}()

	return nil
}

//writeOnS3 func takes the job and session
//that to be written on the amazon S3
func (s *S3) writeOnS3(svc *s3.S3, Job job.Job) {
	defer s.wg.Done()

	var buffer []byte
	var err error

	switch Payload := Job.Payload.(type) {
	case payload.Bytes:
		buffer = Payload
	case payload.Stream:
		buffer, err = ioutil.ReadAll(Payload)
		if err != nil {
			Job.ResponseChan <- response.Error(err)
			return
		}
	}

	filePath, err := s.config.filepath.Evaluate(Job.Data)
	if err != nil {
		Job.ResponseChan <- response.Error(err)
		return
	}

	size := int64(len(buffer))
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s.config.S3Bucket),
		Key:                  aws.String(filePath),
		ACL:                  aws.String(s.config.CannedACL),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ContentEncoding:      aws.String(s.config.Encoding),
		ServerSideEncryption: aws.String(s.config.ServerSideEncryptionAlgorithm),
		StorageClass:         aws.String(s.config.StorageClass),
	})

	if err != nil {
		Job.ResponseChan <- response.Error(err)
		return
	}

	// send response
	Job.ResponseChan <- response.Ack()
}

//Stop func Send a close signal to stop chan
// to stop taking jobs and Stop everything safely
func (s *S3) Stop() error {
	s.wg.Wait()
	return nil
}

func pingBucket(bucket string, client *s3.S3) error {
	tst := s3.GetBucketLoggingInput{Bucket: &bucket}
	_, err := client.GetBucketLogging(&tst)
	return err
}

func getClient(region string, creds *credentials.Credentials) (*s3.S3, error) {
	s3Config := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	sess, err := session.NewSession(s3Config)
	if err != nil {
		return &s3.S3{}, err
	}
	svc := s3.New(sess, s3Config)
	return svc, nil
}

func getCredentials(s *S3) (*credentials.Credentials, error) {
	staticCreds := credentials.NewStaticCredentials(s.config.AccessKeyID, s.config.SecretAccessKey, s.config.SessionToken)
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
	return creds, err
}
