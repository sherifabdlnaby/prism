### label Operation

#### Description

This is the label operation in Vips processing plugin which will put a water mark on the photo.

##### Usage
This is an example of label operation config:

    vips:
        plugin: vips
            config:
                operations:                     
                    label:
                        text: "@TestWatermark"
                        dpi: 100
                        width: 100
                        font: sans 15 bold
                        opacity: 0.75        
    
    
#### Label Operation Configuration Options

This Operation supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [width](#width)  |  integer        | no     | yes     |
| [dpi](#dpi)  |  integer        | no     | yes     |
| [margin](#margin)  |  integer        | no     | yes     |
| [opacity](#opacity)  |  integer        | no     | yes     |
| [replicate](#replicate)  |  boolean        | no     | yes     |
| [text](#text)  |  string        | no     | yes     |
| [font](#font)  |  string        | no     | yes     |
| [colorR](#colorR)  |  integer        | no     | yes     |
| [colorG](#colorG)  |  integer        | no     | yes     |
| [colorB](#colorB)  |  integer       | no     | yes     |


##### `width`
 * This is an optional setting
 * Value type is integer
 * There is default value for this setting is empty
 * This setting supports dynamic values
 * This setting specify the width of the watermark

 ##### `dpi`
  * This is an optional setting
  * Value type is integer
  * The default value for this setting is `10`.
  * This setting supports dynamic values
  * This setting specify the dots per inch for the water mark 
 
 ##### `margin`
   * This is an optional setting
   * Value type is integer
   * The default value for this setting is `0`.
   * This setting supports dynamic values
   * This setting specify the margin for the watermark
  
 ##### `opacity`
   * This is an optional setting
   * Value type is integer
   * The default value for this setting is `5`.
   * This setting supports dynamic values
   * This setting specify the opacity for the watermark
   
 ##### `replicate`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `false`.
   * This setting supports dynamic values
   * This setting specify if the watermark will be repeated or not 

 ##### `text`
   * This is an optional setting
   * Value type is string
   * There is no default value for this setting.
   * This setting supports dynamic values
   * This setting specify what will be written in the watermark
  
 ##### `font`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `sans 10`
   * This setting supports dynamic values
   * This setting specify the font name and it's size
   
 ##### `colorR`
   * This is an optional setting
   * Value type is integer
   * The default value for this setting is `255`
   * This setting supports dynamic values
   * This setting specify the density of the red color in the watermark
   
 ##### `colorG`
   * This is an optional setting
   * Value type is integer
   * The default value for this setting is `255`
   * This setting supports dynamic values
   * This setting specify the density of the green color in the watermark


 ##### `colorB`
   * This is an optional setting
   * Value type is integer
   * The default value for this setting is `255`
   * This setting supports dynamic values
   * This setting specify the density of the blue color in the watermark
