### resize Operation

#### Description

This is the resize operation in Vips processing plugin which will resize the photo.

##### Usage
This is an example of resize operation config:

    vips:
        plugin: vips
            config:
                operations:                     
                    resize:
                        max_width: 800
                        strategy: "embed"  
    
    
#### Crop Operation Configuration Options

This Operation supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [width](#width)  |  integer        | no     | yes     |
| [height](#height)  |  integer        | no     | yes     |
| [max_height](#max_height)  |  integer        | no     | yes     |
| [max_width](#max_width)  |  integer        | no     | yes     |
| [min_height](#min_height)  |  integer        | no     | yes     |
| [min_width](#min_width)  |  integer        | no     | yes     |
| [strategy](#strategy)  |  string        | no     | yes     |


##### `width`
 * This is an optional setting
 * Value type is integer
 * There is no default value for this setting.
 * This setting supports dynamic values

 ##### `height`
  * This is an optional setting
  * Value type is integer
  * There is no default value for this setting.
  * This setting supports dynamic values
 
 ##### `max_height`
   * This is an optional setting
   * Value type is integer
   * There is no default value for this setting.
   * This setting supports dynamic values
  
 ##### `max_width`
   * This is an optional setting
   * Value type is integer
   * There is no default value for this setting.
   * This setting supports dynamic values
   
 ##### `min_height`
   * This is an optional setting
   * Value type is integer
   * There is no default value for this setting.
   * This setting supports dynamic values

 ##### `min_width`
   * This is an optional setting
   * Value type is integer
   * There is no default value for this setting.
   * This setting supports dynamic values
  
 ##### `strategy`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `embed`
   * This setting supports dynamic values
   * The setting must be one of `embed`, `crop`, `stretch`