### Validator processing plugin

#### Description

This plugin checks the type of the file getting processed 
that allows specific file types provided from config
and stops request if its not.

It can also check min/max width and min/max height.

##### Requirements

*At least one format to accept.

##### Usage
This is an example of validator config:

    format: jpeg                                                    (required)
 
    
#### Validator Plugin Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [format](#format)  |  []String        | yes     | no     |
| [maxheight](#maxheight)  |  int        | no     | no     |
| [minheight](#minheight)  |  int        | no     | no     |
| [maxwidth](#maxwidth)  |  int        | no     | no     |
| [minwidth](#minwidth)  |  int        | no     | no     |

##### `format`
 * This is a required setting
 * At least one format should be provided
 * Formats available are {jpeg/jpg, PNG, webp}

##### `maxheight`
   *This is an optional setting
  
   *Max height allowed for an input
 
##### `minheight`
   *This is an optional setting
   
   *Min height allowed for an input
  
##### `maxwidth`
   *This is an optional setting
   
   *Max width allowed for an input 
  
##### `minwidth` 

   *This is an optional setting
   
   *Min width allowed for an input