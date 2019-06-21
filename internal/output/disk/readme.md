### Disk output plugin

#### Description

This plugin saves the output pictures into the disk.

##### Usage
This is an example of disk config:
    
    disk:
        plugin: disk
        concurrency: 100
        config:
            filepath: output/1/disk_output-1-@{count}.jpg          (required)
            permission: 0777                                       (optional)
    
#### Disk Output Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [filepath](#filepath)  |  string        | yes     | yes     |
| [permission](#permission)  |  string            |   no     |   no     |

##### `filepath`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting
 * This setting supports dynamic values
 * This setting specify the path where the picture will be saved, If any of the folders in the path isn't there it will be created.

##### `permission`
 * This is an optional setting
 * Value type is string
 * This setting specify the permission of the picture
 * The default value for this setting is `0777`  (everyone can read, write and execute)
 * Value can be any of: `0000`, `0700`, `0770`, `0777`, `0111`, `0222`, `0333`, `0444`, `0555`, `0666`, `0740`
 
 
