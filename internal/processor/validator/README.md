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
| [maxheight](#maxheight)  |  int        | yes     | no     |
| [minheight](#minheight)  |  int        | yes     | no     |
| [maxwidth](#maxwidth)  |  int        | yes     | no     |
| [minwidth](#minwidth)  |  int        | yes    | no     |

##### `format`
 * This is a required setting
 * At least one format should be provided
 * Formats available are {jpeg/jpg, PNG, webp}

##### `maxheight`
   *Should be 0 or more, 0 for non check.
  
   *Max height allowed for an input
 
##### `minheight`
   *Should be 0 or more, 0 for non check.
   
   *Min height allowed for an input
  
##### `maxwidth`
   *Should be 0 or more, 0 for non check.
   
   *Max width allowed for an input 
  
##### `minwidth` 

   *Should be 0 or more, 0 for non check.
   
   *Min width allowed for an input