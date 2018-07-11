package tests

import (
	"encoding/json"
	"os/exec"
	"testing"

	"flag"

	"os"

	"bytes"

	"strings"

	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
)

// Tests on external application. Need to be generated first.

//go:generate go build -o ../testapp/testapp ../testapp
//go:generate go build -o ../testappConfig/testapp ../testapp
//go:generate go build -o ../testappMulti/testapp ../testappMulti
//go:generate go build -o ../testappMultiConfig/testapp ../testappMulti

type TestConfig struct {
	Port       int
	CpuProfile bool
	DbUri      string
	GrpcUri    string
}

var defaultConf = TestConfig{
	Port:       8080,
	CpuProfile: false,
	DbUri:      "boltdb:db/my.db",
	GrpcUri:    "http://localhost:5432",
}

func TestMainCommandDefaultValues(t *testing.T) {
	actualConf := RunSingle(false)

	assertConf(t, &actualConf, &TestConfig{
		Port:       defaultConf.Port,
		CpuProfile: defaultConf.CpuProfile,
		DbUri:      defaultConf.DbUri,
	})
}

func TestMainCommandFlags(t *testing.T) {
	actualConf := RunSingle(false, "--port=1234", "--cpuprofile=true")

	assertConf(t, &actualConf, &TestConfig{
		Port:       1234,
		CpuProfile: true,
		DbUri:      defaultConf.DbUri,
	})
}

func TestMainCommandDefaultConfigFile(t *testing.T) {
	actualConf := RunSingle(true)

	assertConf(t, &actualConf, &TestConfig{
		Port:       4444,
		CpuProfile: true,
		DbUri:      "inmemory",
	})
}

/*
Precedence is (order or importance):
1) command line flags
2) environment variables
3) config file
4) default values
*/
func TestMainCommandPrecedence(t *testing.T) {
	defer os.Unsetenv("TESTAPP_PORT")
	defer os.Unsetenv("TESTAPP_DBURI")
	os.Setenv("TESTAPP_PORT", "3000")
	os.Setenv("TESTAPP_DBURI", "from_env")

	actualConf := RunSingle(true, "--port=5555")

	assertConf(t, &actualConf, &TestConfig{
		Port:       5555,
		CpuProfile: true, // From config file
		DbUri:      "from_env",
	})
}

func TestMainCommandCustomConfigLocation(t *testing.T) {
	actualConf := RunSingle(true, "--config=config.custom.toml")

	assertConf(t, &actualConf, &TestConfig{
		Port:       9999,
		CpuProfile: true,
		DbUri:      "custom-dburi",
	})
}

func TestMainCommandCustomConfigFromEnv(t *testing.T) {
	defer os.Unsetenv("TESTAPP_CONFIG")
	os.Setenv("TESTAPP_CONFIG", "config.custom.toml")
	actualConf := RunSingle(true)

	assertConf(t, &actualConf, &TestConfig{
		Port:       9999,
		CpuProfile: true,
		DbUri:      "custom-dburi",
	})
}

func TestMainCommandEnvironmentVariables(t *testing.T) {
	defer os.Unsetenv("TESTAPP_PORT")
	defer os.Unsetenv("TESTAPP_DBURI")
	defer os.Unsetenv("TESTAPP_CPUPROFILE")
	os.Setenv("TESTAPP_PORT", "3000")
	os.Setenv("TESTAPP_DBURI", "inmemory")
	os.Setenv("TESTAPP_CPUPROFILE", "true")
	actualConf := RunSingle(false)

	assertConf(t, &actualConf, &TestConfig{
		Port:       3000,
		CpuProfile: true,
		DbUri:      "inmemory",
	})
}

func TestConfigMultipleCommandsDefaultValues(t *testing.T) {
	httpgwConf := RunMulti(false, "httpgw")

	assertConf(t, &httpgwConf, &TestConfig{
		Port:       defaultConf.Port,
		CpuProfile: defaultConf.CpuProfile,
		GrpcUri:    defaultConf.GrpcUri,
	})

	grpcConf := RunMulti(false, "grpc")

	assertConf(t, &grpcConf, &TestConfig{
		Port:       defaultConf.Port,
		CpuProfile: defaultConf.CpuProfile,
		DbUri:      defaultConf.DbUri,
	})
}

func TestConfigMultipleCommandsFlags(t *testing.T) {
	httpgwConf := RunMulti(false, "httpgw", "--port=4000")

	assertConf(t, &httpgwConf, &TestConfig{
		Port:       4000,
		CpuProfile: defaultConf.CpuProfile,
		GrpcUri:    defaultConf.GrpcUri,
	})

	grpcConf := RunMulti(false, "grpc", "--cpuprofile=true", "--dbUri=inmemory:2")

	assertConf(t, &grpcConf, &TestConfig{
		Port:       defaultConf.Port,
		CpuProfile: true,
		DbUri:      "inmemory:2",
	})
}

func TestConfigMultipleCommandsEnvironmentVariables(t *testing.T) {
	defer os.Unsetenv("TESTAPP_HTTPGW_PORT")
	defer os.Unsetenv("TESTAPP_HTTPGW_CPUPROFILE")
	defer os.Unsetenv("TESTAPP_GRPC_PORT")
	defer os.Unsetenv("TESTAPP_GRPC_DBURI")
	os.Setenv("TESTAPP_HTTPGW_PORT", "3000")
	os.Setenv("TESTAPP_HTTPGW_CPUPROFILE", "true")
	os.Setenv("TESTAPP_GRPC_PORT", "4000")
	os.Setenv("TESTAPP_GRPC_DBURI", "inmemory")

	httpgwConf := RunMulti(false, "httpgw")
	assertConf(t, &httpgwConf, &TestConfig{
		Port:       3000,
		CpuProfile: true,
		GrpcUri:    defaultConf.GrpcUri,
	})

	grpcConf := RunMulti(false, "grpc")
	assertConf(t, &grpcConf, &TestConfig{
		Port:       4000,
		CpuProfile: defaultConf.CpuProfile,
		DbUri:      "inmemory",
	})
}

func TestConfigMultipleCommandsConfigFile(t *testing.T) {
	httpgwConf := RunMulti(true, "httpgw")

	assertConf(t, &httpgwConf, &TestConfig{
		Port:       6666,
		CpuProfile: true,
		GrpcUri:    "http://my.friend:5432",
	})

	grpcConf := RunMulti(true, "grpc")

	assertConf(t, &grpcConf, &TestConfig{
		Port:       7777,
		CpuProfile: false,
		DbUri:      "inmemory:multi",
	})
}

func RunSingle(useConfig bool, args ...string) TestConfig {
	return Run(false, useConfig, args...)
}
func RunMulti(useConfig bool, args ...string) TestConfig {
	return Run(true, useConfig, args...)
}
func Run(multi, useConfig bool, args ...string) TestConfig {
	workDir := "../testapp"
	if multi {
		workDir += "Multi"
	}
	if useConfig {
		workDir += "Config"
	}
	return RunTestApp(workDir, args...)
}

func RunTestApp(workDir string, args ...string) TestConfig {
	cmd := exec.Command(workDir+"/testapp", args...)
	cmd.Dir = workDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		glog.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		glog.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		glog.Fatal(err)
	}
	// Read stdoutString
	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	stdoutString := buf.String()

	var testConf TestConfig
	if err := json.NewDecoder(strings.NewReader(stdoutString)).Decode(&testConf); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(stderr)
		stderrString := buf.String()

		glog.Errorf("stdout: " + stdoutString)
		glog.Errorf("stderr: " + stderrString)
		glog.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		glog.Fatal(err)
	}
	return testConf
}

func init() {
	setDefaultFlag("stderrthreshold", "INFO")
	setDefaultFlag("vmodule", "*=0")
	setDefaultFlag("v", "1")
}

func setDefaultFlag(name, value string) {
	existing := flag.Lookup(name)
	if existing == nil {
		flag.Set(name, value)
	}
}

func assertConf(t *testing.T, actualConf, expectedConf *TestConfig) {
	assert.Equal(t, expectedConf.Port, actualConf.Port, "Incorrect port")
	assert.Equal(t, expectedConf.CpuProfile, actualConf.CpuProfile, "Incorrect cpuprofile")
	assert.Equal(t, expectedConf.DbUri, actualConf.DbUri, "Incorrect dbUri")
	assert.Equal(t, expectedConf.GrpcUri, actualConf.GrpcUri, "Incorrect grpcUri")
}
