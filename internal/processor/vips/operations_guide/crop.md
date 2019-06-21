### Crop Operation

#### Description

This is the crop operation in Vips processing plugin which will crop the photo.

##### Usage
This is an example of crop operation config:

    width:  300
    height: 300
    anchor: smart
    
#### Crop Operation Configuration Options

This Operation supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [width](#width)  |  string        | yes     | yes     |
| [height](#height)  |  string        | yes     | yes     |
| [anchor](#anchor)  |  string        | no     | yes     |


***Note: width and height settings are required only if you decided to use the crop operation***

##### `width`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting.
 * This setting supports dynamic values

 ##### `height`
  * This is a required setting
  * Value type is string
  * There is no default value for this setting.
  * This setting supports dynamic values
  
  ##### `anchor`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `center`.
   * This setting supports dynamic values
   * The setting must be one of `center`, `north`, `east`, `south`, `west`, `smart`
