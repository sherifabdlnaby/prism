### HTTP input plugin

#### Description

This plugin starts a webserver and accepts requests based on the config file.

##### Requirements

* A server exposed to the clients requesting.
* A Port to run on.
* Cert and Key files in case of https

##### Usage
 The http plugin starts a web server instance waiting for any request, parses it, checks for errors and passes it to the engine.
 
 Note that for any request, the dynamic values required in config files should be set as parameters
 in the request or the engine will drop the request raising an error that some required value was not set.
 
 eg:
 `{
      "code": 400,
      "message": "error while processing, reason: base [width] is not found in transaction"
  }`

 This is an example of http config:
        
     config:
        port: 80                                                        (required)
        image_field: uploadphoto                                        (optional)
        paths:                                                          (required)
            "/dynamic_resize":
                pipeline: "dynamic_resize"
            "/pipeline":
                pipeline: "@{processing_method}"
        rate_limit: 5                                                   (optional)
        log_requests: all                                               (optional)
        log_errors: false                                               (optional)
 
    
#### Http Plugin Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [port](#port)  |  integer        | yes     | no     |
| [image_field](#image_field)  |  string            |   no     | no     |
| [certFile](#https_config)  | string       |    no     | no     |
| [keyFile](#https_config)  |  string        | no     | no     |
| [paths](#paths)  |  map[string]path            |  yes`validate:"min=1"`     | no     |
| [log_request](#log_request)  | string       |    no`validate:"oneof=all debug none"`    | no     |
| [log_errors](#log_errors)  |  boolean        | no     | no     |
| [rate_limit](#rate_limit)  |  float64            |   no     | no     |
##### `port`
 * This is a required setting
 * Value type is integer
 * There is no default value for this setting.
 * The number provided is what the web server runs on.

##### `image_field`
 * The plugin accepts file upload requests by parsing multipart requests, for an uploaded file
 , the name of the form in request must be the same as one in config.
 * The default value is **image**
 * Value type is string
 
##### `https_config`
  * This is an optional setting, but should be set in order to have https.
  * Value type is string which is the directory for key file.
  * There is no default value for this setting.
  * If none is set, the plugin will run in http mode.
##### `paths`
  
         paths:
                        "/dynamic_resize":
                            pipeline: "dynamic_resize"
                        "/pipeline":
                            pipeline: "@{processing_method}"
                        
  * At least 1 path has to be set.
  * The string `/dynamic_resize` , `/pipeline` are endpoints that the plugin receives requests at.
  * As in the above example, it might take 2 forms:
    * Static: As in `/dynamic_resize`, the photo uploaded will always go through
    `dynamic_resize` pipeline.
    * Dynamic: As in `/pipeline`, the pipeline should be provided as a parameter named `processing_method` in the request
    and not setting it returns an error.
   
  * There is no default value for this setting.  
  
##### `log_request` 

  * Value type is string
  * Can take one of 3 values 
  
    `none`: No requests shall be logged.
    
    `debug`: Only log requests if in debugging mode.
    
    `all`: Always log requests.
##### `log_errors`
  * Value type is boolean
    * `false`: Don't log errors.
    * `true`: Log errors.
        
##### `rate_limit`
  * Value type is Float64.
  * If not set or set by 0, no rate limit.
  * The number set is the number of maximum requests per second.
  
