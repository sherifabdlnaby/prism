### Mysql output plugin

#### Description

This plugin saves picture's data into a mysql database.

##### Usage
This is an example of mysql config:

    mysql:
        plugin: mysql
        concurrency: 100
        config:
            username: root                                                                                                                                       (required)
            db_name: mydatabase                                                                                                                                  (required)        
            query: INSERT INTO photos (ID, photoID, length, width, location) VALUES (@{count},'@{_id}','45','85','Cairo');                                       (required)
    
#### Mysql Output Configuration Options

This plugin supports the following configuration options.

|Setting   |Input type      |  Required |  Dynamic |
|-----------|----------------------|-----------|-----------|
| [username](#username)  |  string        | yes     |   no     |
| [password](#password)  |  string            |   no     |   no     |
| [db_name](#db_name)  |  string        | yes     |   no     |
| [query](#query)  |  string            |   yes     |   yes     |

##### `username`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting
 * This setting specify the username 

##### `password`S
 * This is an optional setting
 * Value type is string
 * The default value for this setting is empty string
 * This setting specify the password


##### `db_name`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting
 * This setting specify the database name

##### `query`
 * This is a required setting
 * Value type is string
 * There is no default value for this setting.
 * This setting specify the query that will be executed
 * This setting supports dynamic values
 * This setting must be a valid mysql query in order to be fulfilled
 * More information and examples for mysql queries from [here](https://dev.mysql.com/doc/mysql-tutorial-excerpt/5.5/en/examples.html)
