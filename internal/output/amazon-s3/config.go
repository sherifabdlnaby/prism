package s3

import cfg "github.com/sherifabdlnaby/prism/pkg/config"

//config struct
type config struct {
	FilePath                      string `mapstructure:"filepath" validate:"required"`
	S3Region                      string `mapstructure:"s3_region" validate:"required"`
	S3Bucket                      string `mapstructure:"s3_bucket" validate:"required"`
	AccessKeyID                   string `mapstructure:"access_key_id"`
	SecretAccessKey               string `mapstructure:"secret_access_key"`
	SessionToken                  string `mapstructure:"session_token"`
	CannedACL                     string `mapstructure:"canned_acl" validate:"oneof=private public-read public-read-write authenticated-read aws-exec-read bucket-owner-read bucket-owner-full-control log-delivery-write"`
	Encoding                      string `mapstructure:"encoding" validate:"oneof=none gzip"`
	ServerSideEncryptionAlgorithm string `mapstructure:"server_side_encryption_algorithm" validate:"oneof=AES256 aws:kms"`
	StorageClass                  string `mapstructure:"storage_class" validate:"oneof=STANDARD REDUCED_REDUNDANCY STANDARD_IA"`

	filepath cfg.Selector
}

//defaultConfig func return the default configurations
func defaultConfig() *config {
	return &config{
		CannedACL:                     "private",
		Encoding:                      "none",
		ServerSideEncryptionAlgorithm: "AES256",
		StorageClass:                  "STANDARD",
	}
}
