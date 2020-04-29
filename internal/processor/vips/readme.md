### VIPS processing plugin

#### Description

This plugin can do some operations on the picture and exports it with configurable format,compression and quality 

##### Usage
This is an example of vips config:

    vips:
        plugin: vips
        config:
            operations:                     (required)
                blur:
                    sigma:  2
                    min_ampl: 10
                flip:
                    direction: both
                resize:
                    max_width: 800
                    strategy: "embed"
            export:                         (required)
                format: jpeg
                quality: 90
    
#### VIPS Output Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [operations](#operations)  |  config        | yes     | no     |
| [export](#export)  |  config            |   yes     | no     |


##### `operations`
 * This is a required setting
 * Value type is config
 * The default value for this setting will do no operation.
 * The operations can be one or more from the below 
    * [blur](./operations_guide/blur.md)
    * [flip](./operations_guide/flip.md)
    * [rotate](./operations_guide/rotate.md)
    * [crop](./operations_guide/crop.md)
    * [resize](./operations_guide/resize.md)
    * [label](./operations_guide/label.md)

 
##### `export`
 * This is a required setting
 * Value type is config
 * This setting controls the export of the picture
 * There is a default config for this setting which is
      
       format: jpeg
       extend: black
       quality: 85
       compression: 6
       strip_metadata: true
       
 export supports the following configuration options
 
 |Setting   |Input type      |  Required |  Must be       |  Default  |
 |-----------|----------------------|-----------|-----------|-----------|
 | format |  string        | no     | jpg, jpeg, png, webp    |jpeg     |
 | extend |  string            |   no     | black, copy, repeat, mirror, white, last     |black     |
 | quality |  int        | no     | between 1 and 100     |85     |
 | compression |  int            |   no     | between 1 and 9     |6     |
 | strip_metadata |  boolean        | no     | true or false     |true     |
 | interlace |  boolean            |   no     | true or false     | true
