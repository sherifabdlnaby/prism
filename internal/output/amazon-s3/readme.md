### S3 output plugin

#### Description

This plugin uploads the output pictures into Amazon Simple Storage Service (Amazon S3).

##### Requirements

* Amazon S3 Bucket and S3 Access Permissions (Typically access_key_id and secret_access_key)
* S3 PutObject permission

##### Usage
This is an example of s3 config:

    filepath: s3_output-1-@{count}-x-@{_timestamp}-x-@{_id}.jpg     (required)
    s3_region: us-east-2                                            (required)
    s3_bucket: prism.test                                           (required)
    access_key_id: "your_access_key_here"                           (optional)
    secret_access_key: "your_secret_access_key_here"                (optional)
    session_token: "your_session_token"                             (optional)
    
#### S3 Output Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [filepath](#filepath)  |  string        | yes     | yes     |
| [s3_region](#s3_region)  |  string            |   yes     | no     |
| [s3_bucket](s3_bucket)  | string       |    yes     | no     |
| [access_key_id](#access_key_id)  |  string        | no     | no     |
| [secret_access_key](#secret_access_key)  |  string            |   no     | no     |
| [session_token](#session_token)  | string       |    no     | no     |
| [canned_acl](#canned_acl)  |  string        | no     | no     |
| [encoding](#encoding)  |  string            |   no     | no     |
| [server_side_encryption_algorithm](#server_side_encryption_algorithm)  | string       |    no     | no     |
| [storage_class](#storage_class)  | string       |    no     | no     |

##### `filepath`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting.
 * This setting supports dynamic values

##### `s3_region`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting.
 
##### `s3_bucket`
  * This is a required setting
  * Value type is string
  * There is no default value for this setting.

##### `access_key_id` 

  * Value type is string
  * There is no default value for this setting.

This plugin uses the AWS SDK and supports several ways to get credentials, which will be tried in this order:

* Static configuration, using `access_key_id` and `secret_access_key` params in logstash plugin config
* Environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
* External credentials file specified by `aws_credentials_file`

##### `secret_access_key`
  * Value type is string
  * There is no default value for this setting.

##### `session_token`
The AWS Session token for temporary credential
  * Value type is string
  * There is no default value for this setting.
  
##### `canned_acl`
The S3 canned ACL to use when putting the photo.
  * Value can be any of: `private`, `public-read`, `public-read-write`, `authenticated-read`, `aws-exec-read`, `bucket-owner-read`, `bucket-owner-full-control`
  * Default value is `private`
  
##### `encoding`
Specify the content encoding. 
  * Value can be any of: `none`, `gzip`
  * Default value is `none`

##### `server_side_encryption_algorithm`
Specifies what type of encryption to use.
  * Value can be any of: `AES256`, `aws:kms`
  * Default value is `AES256`

##### `storage_class`
Specifies what S3 storage class to use when uploading the photo.
  * Value can be any of: `STANDARD`, `REDUCED_REDUNDANCY`, `STANDARD_IA`
  * Default value is `STANDARD`

