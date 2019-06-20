### Nudity detector processing plugin

#### Description

This plugin checks he likely hood that this image contains nudity 
and will censor the image with configured color pixels, or 
it can be configured to send a NoAck when it detects nudity or just 
add a flag to payload.Data

##### Requirements

*At least one format to accept.

##### Usage
This is an example of Nudity detector config:

    format: jpeg                                                    (required)
 
    
#### Nudity detector Plugin Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [format](#format)  |  []String        | yes     | no     |
| [drop](#drop)  |  bool        | no     | no     |
| [R](#R)  |  uint8        | no     | no     |
| [G](#G)  |  uint8        | no     | no     |
| [B](#B)  |  uint8        | no     | no     |
| [A](#A)  |  float64        | no     | no     |
| [quality](#quality)  |  int        | yes     | no     |

##### `format`
 * This is a required setting
 * At least one format should be provided
 * Formats available are {jpeg/jpg, PNG}

##### `Quality`
   *This is a required setting
   *From 1 to 100 
 
##### `G`
   *This is a required setting
   
   *The intensity of green
  
##### `R`
   *This is a required setting
   
   *The intensity of red
  
##### `B` 

   *This is a required setting
   
   *The intensity of Blue
   
##### `A` 

   *This is a required setting
   
   *The opacity of the color
