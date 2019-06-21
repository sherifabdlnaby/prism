### label Operation

#### Description

This is the label operation in Vips processing plugin which will put a water mark on the photo.

##### Usage
This is an example of label operation config:

    text: "@TestWatermark"
    dpi: 100
    width: 100
    font: sans 15 bold
    opacity: 0.75
    
#### Label Operation Configuration Options

This Operation supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [width](#width)  |  string        | no     | yes     |
| [dpi](#dpi)  |  string        | no     | yes     |
| [margin](#margin)  |  string        | no     | yes     |
| [opacity](#opacity)  |  string        | no     | yes     |
| [replicate](#replicate)  |  string        | no     | yes     |
| [text](#text)  |  string        | no     | yes     |
| [font](#font)  |  string        | no     | yes     |
| [colorR](#colorR)  |  string        | no     | yes     |
| [colorG](#colorG)  |  string        | no     | yes     |
| [colorB](#colorB)  |  string        | no     | yes     |


##### `width`
 * This is an optional setting
 * Value type is string
 * There is no default value for this setting.
 * This setting supports dynamic values

 ##### `dpi`
  * This is an optional setting
  * Value type is string
  * The default value for this setting is `10`.
  * This setting supports dynamic values
 
 ##### `margin`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `0`.
   * This setting supports dynamic values
  
 ##### `opacity`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `5`.
   * This setting supports dynamic values
   
 ##### `replicate`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `false`.
   * This setting supports dynamic values

 ##### `text`
   * This is an optional setting
   * Value type is string
   * There is no default value for this setting.
   * This setting supports dynamic values
  
 ##### `font`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `sans 10`
   * This setting supports dynamic values
   
 ##### `colorR`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `255`
   * This setting supports dynamic values
   
 ##### `colorG`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `255`
   * This setting supports dynamic values

 ##### `colorB`
   * This is an optional setting
   * Value type is string
   * The default value for this setting is `255`
   * This setting supports dynamic values