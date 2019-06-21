### Crop Operation

#### Description

This is the crop operation in Vips processing plugin which will crop the photo.

##### Usage
This is an example of crop operation config:

    vips:
        plugin: vips
            config:
                operations:                     
                    crop:
                        width:  300
                        height: 300
                        anchor: smart
                            
    
    
#### Crop Operation Configuration Options

This Operation supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [width](#width)  |  integer       | yes     | yes     |
| [height](#height)  |  integer        | yes     | yes     |
| [anchor](#anchor)  |  string        | no     | yes     |


***Note: width and height settings are required only if you decided to use the crop operation***

##### `width`
 * This is a required setting
 * Value type is integer
 * There is no default value for this setting
 * This setting supports dynamic values
 * This setting specify the width of the picture that will be cropped

 ##### `height`
  * This is a required setting
  * Value type is integer
  * There is no default value for this setting.
  * This setting supports dynamic values
  * This setting specify the height of the picture that will be cropped
  
  ##### `anchor`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `center`
   * This setting supports dynamic values
   * The setting must be one of `center`, `north`, `east`, `south`, `west`, `smart`
   * This setting specify where the crop will be done, the `smart` setting will choose the area with the max variants (mostly the faces) and crop it.
