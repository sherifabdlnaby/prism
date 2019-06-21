### Rotate Operation

#### Description

This is the rotate operation in Vips processing plugin which will rotate the photo.

##### Usage
This is an example of rotate operation config:

    vips:
        plugin: vips
            config:
                operations:                     
                    rotate:
                        angle: 90

#### Rotate Operation Configuration Options

This Operation supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [angle](#angle)  |  integer        | yes     | yes     |

***Note: The angle setting is required only if you decided to use the rotate operation***

##### `angle`
 * This is a required setting
 * Value type is integer
 * There is no default value for this setting.
 * This setting supports dynamic values
 * The setting must be one of `0`, `90`, `180`, `270` 
 * This setting specify the angle of the rotation

 
