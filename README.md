# Pungi
Pungi's goal is to simplify command line application startup and configuration. Pungi is built upon [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper), hiding their complexity from the user. Simplicity comes from exposing less functionality and forcing some standards.

## Features
* Supports commands
* Configuration overloading
* Configuration definition in Go code

## Hello World Example
```go
package main
import "github.com/joosep-wm/pungi"

func main() {
  pungi.New("testapp", "Prints hello <your name>.").
    Key("name", "John", "The name to print").
    Run(startApp).
    Execute()
}

func startApp(conf *pungi.Conf, args []string) error {
  println("Hello " + conf.GetString("name"))
  return nil
}
```
In this example the application `testapp` has one string typed configuration parameter: `name`.

## Using Commands 
Application can use commands for different behaviour.
For example:
```go
package main
import "github.com/joosep-wm/pungi"

func main() {
  pungi.New("testapp", "Starts music store web application.").
    Key("cpuprofile", false, "Starts CPU profiler if set to true.").
    Cmd(pungi.Cmd("grpc", "Starts gRPC service.", startGrpcService).
      Key("port", 5432, "Service listen port.").
      Key("dbUri", "boltdb:db/my.db", "Db Uri"),
    ).
    Cmd(pungi.Cmd("httpgw", "Starts Http GW.", startHttpGW).
      Key("port", 8080, "Http GW listen port.").
      Key("grpcUri", "http://localhost:5432", "Grpc service Uri."),
    ).
    Execute()
}
func startHttpGW(conf *pungi.Conf, args []string) error {
  // GW start code here
  return nil
}

func startGrpcService(conf *pungi.Conf, args []string) error {
  // GRPC start code here
    return nil
}
```
In this example the configuration key `cpuprofile` is used by both commands. Each command has their own specific configuration values.

## Configuration Key Lookup
Configuration values are looked up in the following order:  
1. Command line flags
2. Environment variables
3. Configuration file
4. Default values

### Use command line flags
Command line flags overload all other sources of configuration. Some examples:
* `testapp grpc --cpuprofile=true --port=4444`
* `testapp httpgw --port=8000`

### Use Environment variables
Environment variables use this naming convention: 
* APPNAME_KEY - for global configuration keys. E.g. `TESTAPP_CONFIG`
* APPNAME_CMD_KEY - for command specific configuration keys. E.g. `TESTAPP_HTTPGW_PORT`

### Use configuration file
The configuration file uses [TOML](https://github.com/toml-lang/toml) syntax.
Example config file:
```toml
[testapp]
port = 4444
dbUri = "inmemory"
cpuprofile = true

[testapp.httpgw]
port = 6666
cpuprofile = false
grpcUri = "http://my.friend:5432"
```
The `testapp` section can be used to define common configuration values. Sub sections `testapp.httpgw` override the default values.  

### Configuration file location
By default Pungi looks for `config.toml` from the working directory.

There is a special configuration key `config` that can be used to define the file location.

From command line: `--config=config.custom.toml`

From env variables: `export TESTAPP_CONFIG=config.custom.toml` 

From Go code you can use `DefaultConfigFile(filename string)` function when building Pungi.

### Configuration types
The types of configuration objects are taken from the default values. Currently these types are supported:
* int
* string
* bool
* float64 

## Pungi low level features
The most common way to initialize the Pungi is to build the configuration and call `Execute()`. It's also possible to call `Initialize()` instead. This returns a `Pungi` struct.

It exposes the config file used and the config objects (root config and one for each command). 

Also a `Execute(args ...string) error` method is exposed, so the application can be started by it. 