### Nudity detector processing plugin

#### Description

This plugin checks the likely hood that this image contains nudity 
and will censor the image with configured color pixels, or 
it can be configured to drop the request when it detects nudity or just 
adds a flag of what has been detected.

    
##### Usage

 This is an example of nude censorer/detector config:
        
     config:
        Drop: true                                                      (optional)
        Export:                                                         (optional)
              Format:   jpeg
              Quality:  90
        RGBA:                                                           (optional)
                R:90
                G:255
                B:255
                A:0.3
    
 The difference between the censorer and detector is that the detector doesn't take RGBA as config input
 as it only detects and can drop, but doesn't do any processing on photos.
#### Nudity detector/censorer Plugin Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [format](#export)  |  []String        | no     | no     |
| [quality](#export)  |  int        | no     | no     |
| [Drop](#Drop)  |  bool        | no     | no     |
| [R](#R)  |  uint8        | no     | no     |
| [G](#G)  |  uint8        | no     | no     |
| [B](#B)  |  uint8        | no     | no     |
| [A](#Alpha)  |  float64        | no `validate:"min=0,max=1"`    | no     |

##### `export`

         Export:                                                         
                      Format:   jpeg
                      Quality:  90
 * Every processor plugin that outputs a photo should have this property.
 * Formats represents the image format produced by this plugin.
 * Quality can take values from 1 - 100


##### `Drop`
 * Value type is boolean
    * `true` : Drops the transaction and stops the request.
    * `false`: Doesn't drop the transaction and just adds a flag.

##### `G`
 * The intensity of green
  
##### `R`
 * The intensity of red
  
##### `B` 
 * The intensity of Blue
   
##### `Alpha` 
 * The opacity of the color
