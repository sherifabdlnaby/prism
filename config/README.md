# `/configs`
Prism provides an easy way to set up the available plugins and workflow by simply changing text in
a file taking the [yaml](https://yaml.org) format. 
 
In order to get Prism up and running there are five configuration files that should exist.
* [inputs.yaml](#inputs.yaml)
* [processors.yaml](#processors.yaml)
* [outputs.yaml](#outputs.yaml)
* [pipelines.yaml](#pipelines.yaml)
* [prism.yaml](#prism)

An example of the files exist [here](./example)



### inputs.yaml 

This file contains the input plugins that you would like to start and their configuration.

           
 eg:
 
    inputs:
        http_nonsecured:
           plugin: http
           config:
               port: 80
               form_name: image
               paths:
                   "/profile_picture":
                       pipeline: "profile_pic_pipeline"
               certfile: ""
               keyfile: ""
               log_request: "debug"
        https_secured:
           plugin: http
           config:
               port: 8081
               form_name: image
               paths:
                   "/dynamic_resize":
                       pipeline: "dynamic_resize"
                   "/pipeline":
                       pipeline: "@{pipeline}"
               certfile: "${CertDirectoryFile}"
               keyfile: "${KeyDirectoryFile}"
               log_request: "debug"
           
The file starts with `inputs:` and then inside it are the names of the input plugins you want to add,
you may add more than one instance of the same plugin by simply adding more than one in the configuration file
and setting the suitable configuration.


For every added plugin you first add a name as `https_secured` then you choose the plugin
and set the suitable config.
        
    https_secured:
               plugin: http
               config:
               

Inside every plugin, you should find a documentation that provides an explanation of the plugin
and its configuration.



### processors.yaml

This file contains the processor plugins.


eg:

    processors:
        watermark_label:
                concurrency: 4
                plugin: vips
                config:
                    operations:
                        label:
                            text: "@TestWatermark"
                            dpi: 100
                            width: 100
                            font: sans 15 bold
                            opacity: 0.75
                    export:
                        format: jpeg
                        quality: 90
        validator:
                concurrency: 100
                plugin: validator
                config:
                    format:
                        - jpeg
                        - png
                        - webp


  The file starts with `processors:` and then inside it are the names of the processor plugins you want to add,
  you may add more than one instance of the same plugin by simply adding more than one in the configuration file
  and setting the suitable configuration.
  For every added plugin you first add a name as `watermark_label` then you choose the plugin,
  set the suitable concurrency,and set the suitable config.
          
      watermark_label:
                      concurrency: 4
                      plugin: vips
                      config:
                          operations:
                                label:
                                    text: "@TestWatermark"
                                    dpi: 100
                                    width: 100
                                    font: sans 15 bold
                                    opacity: 0.75
                                export:
                                    format: jpeg
                                    quality: 90
                 
  
  
  For processor plugins, some may produce a photo as in `watermark_label` and some may not as in 
  `validator`, so to control the produced photo an `export` field must be added, this is explained more 
  in every processor plugin.
  
  
  Inside every plugin, you should find a documentation that provides an explanation of the plugin
  and its configuration.
  
  


### outputs.yaml
This file contains the output plugins.


eg:

    outputs:
        dynamic:
            plugin: disk
            concurrency: 100
            config:
                filepath: output/dynamic/@{_timestamp}-@{_filename}-@{_width}x@{_height}.@{_format}
                permission: 0777
        original:
            plugin: disk
            concurrency: 100
            config:
                filepath: output/original/@{_timestamp}-@{_filename}.@{_format}
                permission: 0777
            

  The file starts with `outputs:` and then inside it are the names of the output plugins you want to add,
  you may add more than one instance of the same plugin by simply adding more than one in the configuration file
  and setting the suitable configuration.
  For every added plugin you first add a name as `original` then you choose the plugin,
  set the suitable concurrency,and set the suitable config.
          
      original:
                  plugin: disk
                  concurrency: 100
                  config:
                      filepath: output/original/@{_timestamp}-@{_filename}.@{_format}
                      permission: 0777
                      
                 
  
  Inside every plugin, you should find a documentation that provides an explanation of the plugin
  and its configuration.



<!--####`prism.yaml`  to be added -->

### pipelines.yaml
Unlike the other files, this one isn't responsible for plugins initiation and configuration setting, but is 
the most important  as it controls the flow of a photo since it was inputted till its saved.


eg:
   
    pipelines:
        profile_pic_pipeline:
            concurrency: 50
            pipeline:
                validator:
                    next:
                        resize_minimum:
                            next:
                                resize_big:
                                    async: false
                                    next:
                                        big:
                                            async: false
                                resize_medium:
                                    async: false
                                    next:
                                        smart_crop_thumbnail:
                                            next:
                                                thumbnail:
                                                    async: false
                                        watermark_label:
                                            next:
                                                watermark:
                                                    async: false
                                        medium:
                                            async: false
                                resize_small:
                                    async: false
                                    next:
                                        flip_blur:
                                            next:
                                                flipped_blurred:
                                                    async: false
                                        small:
                                            async: false
                                original:
                                    async: false
        dynamic_resize:
            concurrency: 50
            pipeline:
                validate_size:
                    next:
                        dynamic_resize:
                            next:
                                dynamic:
                                    async: false
   


The file starts with `pipelines:` where every pipeline has a name that you provide as `profile_pic_pipeline` 
and a concurrency level and the pipeline itself.

A pipeline may be thought of as a tree consisting of nodes where each plugin is a node, a node's collection of 
children are given by a `next`, where children themselves are executed concurrently and parent is executed
 before its children.
 
 eg:
    
         pipeline:
                validator:
                    next:
                        resize_minimum:
                            next:
                                resize_big:
                                    async: false
                                    next:
                                        big:
                                            async: false
                                resize_small:
                                    async: false
                                    next:
                                        flip_blur:
                                            next:
                                                flipped_blurred:
                                                    async: false
                                        small:
                                            async: false
                                            
As we can see here, `validator` plugin has a child `resize_minimum` plugin, and therefore in execution, `resize_minimum`
plugin won't start its job until `validator` finishes and gives the command for `resize_minimum`.

and in case of `resize_minimum`, it has 2 children `resize_big` and `resize_small`, both will run concurrently
when `resize_minimum` has finished its job.


You can notice that every plugin has a field `async : false` ; normally for a request to be declared successful, 
it has to return an acknowledgment to the previous parent till the root, for `async : true` a child tells the parent
plugin that it shouldn't wait for an acknowledgment.




### Dynamic Values
In order to make the configuration files more dynamic, Prism provides different ways to 
set field values.

Field values can take 3 forms:
        
   eg:     
          
      https_secured:
               plugin: http
               config:
                   port: 8081
                   form_name: image
                   paths:
                       "/dynamic_resize":
                           pipeline: "dynamic_resize"
                       "/pipeline":
                           pipeline: "@{pipeline}"
                   certfile: "${CertDirectoryFile}"
                   keyfile: "${KeyDirectoryFile}"
                   log_request: "debug"
                       
   * Static:
            `port: 8081`
     
     This type is provided in the config file directly and
     will stay the same during the whole execution time.
   * Dynamic:
            `pipeline: "@{pipeline}"`
            
     This type gets evaluated during the execution for example 
     provided by every  request.
   
   * Environment Variables:
            `certfile: "${CertDirectoryFile}"`
            
     This type gets evaluted from the environment variables.
