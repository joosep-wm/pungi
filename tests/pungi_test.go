package tests

import (
	"testing"

	"github.com/joosep-wm/pungi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// In separate package, so external API usage could be seen
func TestTwoCommandsInit(t *testing.T) {
	viper.Reset()
	p, err := pungi.New("musicstore", "Music store web application").
		DefaultConfigFile("config/mstore.toml").
		Key("cpuprofile", true, "Starts CPU profiler if set to true.").
		Cmd(pungi.Cmd("webapp", "Starts web application.", grpcFunc).
			Key("port", 8080, "Listen port for web application."),
		).
		Cmd(pungi.Cmd("grpc", "Starts gRPC service.", grpcFunc).
			Key("port", 5432, "gRPC service listen port.").
			Key("dbUri", "boltdb:db/my.db", "DB Uri"),
		).
		Cmd(pungi.Cmd("httpgw", "Starts HTTP gateway for gRPC service.", httpgwFunc).
			Key("port", 9090, "Http gateway listen port.").
			Key("grpcUrl", "http://localhost:5432", "gRPC service url"),
		).Initialize()

	require.NoError(t, err)

	webConf := p.Config("webapp")
	assert.Equal(t, 8080, webConf.GetInt("port"))
	grpcConf := p.Config("grpc")
	assert.Equal(t, 5432, grpcConf.GetInt("port"))
	hgwConf := p.Config("httpgw")
	assert.Equal(t, 9090, hgwConf.GetInt("port"))
	assert.Equal(t, true, hgwConf.GetBool("cpuprofile"))

	err = p.Execute("webapp")
	require.NoError(t, err)
	assert.Equal(t, "config/mstore.toml", p.ConfigFileUsed())
}

func TestDefaultCommandInit(t *testing.T) {
	viper.Reset()
	p, err := pungi.New("musicstore", "Starts music store web application.").
		Key("port", 8080, "Listen port").
		Key("cpuprofile", false, "Starts CPU profiler if set to true.").
		Run(startWebApp).Initialize()

	require.NoError(t, err)

	conf := p.RootConfig()
	assert.Equal(t, 8080, conf.GetInt("port"))
	assert.Equal(t, false, conf.GetBool("cpuprofile"))

	err = p.Execute("musicstore")
	require.NoError(t, err)
	assert.Equal(t, "config.toml", p.ConfigFileUsed())
}

func TestMainRunnableAndSubcommandsReturnError(t *testing.T) {
	viper.Reset()
	p, err := pungi.New("musicstore", "Starts music store web application.").
		Run(startWebApp).
		Cmd(pungi.Cmd("grpc", "Starts gRPC service", grpcFunc)).
		Initialize()

	require.Error(t, err)
	require.Nil(t, p)
}

func TestMainRunnableAndSubcommandsWorksWithDefinedArgs(t *testing.T) {
	viper.Reset()

	defer func() {
		webAppArgs = []string{}
		grpcArgs = []string{}
	}()

	p, err := pungi.New("musicstore <your-name>", "Main command starts webapp").
		Run(startWebApp).
		Key("port", 8080, "Listen port").
		Args(cobra.ExactArgs(1)).
		Cmd(pungi.Cmd("grpc <your-name>", "Starts gRPC service.", grpcFunc)).
		Initialize()

	require.NoError(t, err)

	err = p.Execute("Joosep")
	require.NoError(t, err)
	assert.Equal(t, []string{"Joosep"}, webAppArgs)

	err = p.Execute("grpc", "Simm")
	assert.Equal(t, []string{"Simm"}, grpcArgs)
	require.NoError(t, err)

	assert.Equal(t, 8080, p.RootConfig().GetInt("port"))
	assert.Equal(t, 8080, p.Config("grpc").GetInt("port"))
}

func TestInvalidKeyValue(t *testing.T) {
	viper.Reset()
	var ohmy struct{}
	p, err := pungi.New("musicstore", "Starts music store web application.").
		Key("port", ohmy, "Listen port").
		Run(startWebApp).
		Initialize()

	require.Error(t, err)
	require.Nil(t, p)
}

func TestFloatKeyValue(t *testing.T) {
	viper.Reset()
	p, err := pungi.New("musicstore", "Starts music store web application.").
		Key("port", float64(2.0), "Listen port").
		Run(startWebApp).
		Initialize()

	require.NoError(t, err)
	require.Equal(t, float64(2.0), p.RootConfig().GetFloat64("port"))
}

func TestOnlyConfigFile(t *testing.T) {
	viper.Reset()
	p, err := pungi.NewConfigFileOnly("testapp", "../testappConfig/config.custom.toml")
	require.NoError(t, err)
	assert.Equal(t, 9999, p.RootConfig().GetInt("port"))
	assert.Equal(t, "custom-dburi", p.RootConfig().GetString("dbUri"))
	assert.Equal(t, true, p.RootConfig().GetBool("cpuprofile"))
}

func TestOnlyConfigFileSetupMulti(t *testing.T) {
	viper.Reset()
	p, err := pungi.NewConfigFileOnly("testapp", "../testappMultiConfig/config.toml")
	require.NoError(t, err)
	assert.Equal(t, 6666, p.Config("httpgw").GetInt("port"))
	assert.Equal(t, 7777, p.Config("grpc").GetInt("port"))
}

func TestAllValues(t *testing.T) {
	viper.Reset()
	p, err := pungi.New("musicstore", "Starts music store web application.").
		Key("port", 8080, "Listen port").
		Key("cpuprofile", false, "Starts CPU profiler if set to true.").
		Run(startWebApp).Initialize()
	require.NoError(t, err)

	err = p.Execute()
	require.NoError(t, err)

	allValues := p.RootConfig().AllValues()
	assert.Equal(t, 8080, allValues["port"])
}

func TestUsageText(t *testing.T) {
	viper.Reset()
	defer func() {
		webAppArgs = []string{}
	}()

	p, err := pungi.New("musicstore <your-name>", "Starts music store web application.").
		Args(cobra.ExactArgs(1)).
		Key("port", 8080, "Listen port").
		Run(startWebApp).Initialize()
	require.NoError(t, err)

	err = p.Execute()
	require.Error(t, err, "Invalid number of arugments should return an error")

	err = p.Execute("Joosep Simm")
	require.NoError(t, err)

	assert.Equal(t, []string{"Joosep Simm"}, webAppArgs)
	assert.Equal(t, 8080, p.RootConfig().GetInt("port"))
}

func TestUsageTextSubCommand(t *testing.T) {
	viper.Reset()
	defer func() {
		webAppArgs = []string{}
	}()

	p, err := pungi.New("musicstore", "Starts music store web application.").
		Cmd(pungi.Cmd("webapp <your-name>", "Starts webapp with your name", startWebApp).
			Args(cobra.MinimumNArgs(1)).
			Key("port", 8080, "Listen port"),
		).
		Initialize()
	require.NoError(t, err)

	err = p.Execute("webapp")
	require.Error(t, err, "Invalid number of arugments should return an error")

	err = p.Execute("webapp", "Joosep Simm")
	require.NoError(t, err)

	assert.Equal(t, []string{"Joosep Simm"}, webAppArgs)
	assert.Equal(t, 8080, p.Config("webapp").GetInt("port"))
}

func TestTwoCommandsAllValues(t *testing.T) {
	viper.Reset()
	p, err := pungi.New("musicstore", "Music store web application").
		DefaultConfigFile("config/mstore.toml").
		Key("cpuprofile", true, "Starts CPU profiler if set to true.").
		Cmd(pungi.Cmd("webapp", "Starts web application.", grpcFunc).
			Key("port", 8080, "Listen port for web application."),
		).
		Cmd(pungi.Cmd("grpc", "Start gRPC service.", grpcFunc).
			Key("port", 5432, "gRPC service listen port.").
			Key("dbUri", "boltdb:db/my.db", "DB Uri"),
		).
		Cmd(pungi.Cmd("httpgw", "Start HTTP gateway for gRPC service.", httpgwFunc).
			Key("port", 9090, "Http gateway listen port.").
			Key("grpcUrl", "http://localhost:5432", "gRPC service url"),
		).Initialize()

	require.NoError(t, err)

	webConf := p.Config("webapp")
	webappSub := webConf.AllValues()

	assert.Equal(t, 8080, webappSub["port"])
	assert.Equal(t, true, webappSub["cpuprofile"])

}

var webAppArgs, grpcArgs, httpArgs []string

func startWebApp(_ *pungi.Conf, args []string) error {
	webAppArgs = args
	return nil
}
func httpgwFunc(_ *pungi.Conf, args []string) error {
	httpArgs = args
	return nil
}
func grpcFunc(_ *pungi.Conf, args []string) error {
	grpcArgs = args
	return nil
}
