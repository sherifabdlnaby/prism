### Flip Operation

#### Description

This is the flip operation in Vips processing plugin which will flip the photo.

##### Usage
This is an example of flip operation config:

    direction: both
    
#### Flip Operation Configuration Options

This Operation supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [direction](#direction)  |  string        | yes     | yes     |

***Note: The direction setting is required only if you decided to use the flip operation***


##### `direction`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting.
 * This setting supports dynamic values
 * The setting must be one of `horizontal`, `vertical`, `both`, `none` 

