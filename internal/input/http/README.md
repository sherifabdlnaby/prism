### HTTP input plugin

#### Description

This plugin starts a webserver and accepts requests based on the config file.

##### Requirements

* A server exposed to the clients requesting.
* A Port to run on.
* Cert and Key files in case of https

##### Usage
This is an example of http config:

    port: 80                                                        (required)
    form_name: image                                                (required)
    paths:                                                          (required)
        "/profile_picture":
            pipeline: "@{pipeline}"
        "/cover_picture/":
            pipeline: "@{pipeline}"
    Ratelimit: 5                                                    (optional)
 
    
#### Http Plugin Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [port](#port)  |  integer        | yes     | no     |
| [form_name](#form_name)  |  string            |   yes     | no     |
| [certFile](#https_config)  | string       |    no     | no     |
| [keyFile](#https_config)  |  string        | no     | no     |
| [paths](#paths)  |  string            |   no     | no     |
| [logrequest](#logrequest)  | integer       |    no     | no     |
| [logresponse](#logresponse)  |  integer        | no     | no     |
| [ratelimit](#ratelimit)  |  float64            |   no     | no     |
##### `port`
 * This is a required setting
 * Value type is integer
 * There is no default value for this setting.

##### `form_name`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting.
 
##### `https_config`
  * This is an optional setting, but should be set in order to have https.
  * Value type is string which is the directory for key file.
  * There is no default value for this setting.
  
##### `paths`
  * At least 1 path has to be set.
  * Every set path should be string and has an inside value which is a pipeline which is dynamic.
  * There is no default value for this setting.  
  
##### `logrequest` 

  * Value type is integer
  * Can take one of 3 values 
  
    0: No requests shall be logged.
    
    1: All requests shall be logged by Debug logger.
    
    2: All requests shall be logged by Info logger.
##### `logresponse`
  * Value type is integer
  * Can take one of 3 values 
  
    0: Nothing shall be logged.
    
    1: All successful requests shall be logged.
    
    2: All Failed requests shall be logged with their errors.
##### `ratelimit`
  * Value type is Float64.
  * If not set or set by 0, no rate limit.
  * The number set is the number of maximum requests per second.
  
