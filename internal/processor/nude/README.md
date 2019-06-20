### Nudity detector processing plugin

#### Description

This plugin checks the likely hood that this image contains nudity 
and will censor the image with configured color pixels, or 
it can be configured to send a NoAck when it detects nudity or just 
add a flag to payload.Data

    
#### Nudity detector Plugin Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [format](#format)  |  []String        | no     | no     |
| [Drop](#Drop)  |  bool        | no     | yes     |
| [R](#R)  |  uint8        | no     | yes     |
| [G](#G)  |  uint8        | no     | yes     |
| [B](#B)  |  uint8        | no     | yes     |
| [A](#A)  |  float64        | no     | yes     |
| [quality](#quality)  |  int        | no     | yes     |

##### `format`
 * At least one format should be provided
 * Formats available are {jpeg/jpg, PNG}

##### `Quality`
   *From 1 to 100 

##### `Drop`
   *Send NoAck to Drop the image

##### `G`
   *The intensity of green
  
##### `R`
   *The intensity of red
  
##### `B` 
   *The intensity of Blue
   
##### `A` 
   *The opacity of the color
