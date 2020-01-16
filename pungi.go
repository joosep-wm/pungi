package pungi

import (
	goflag "flag"
	"fmt"
	"os"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const rootKey = "_root_"

/*
 Start here.

 Define the app configuration and then call `Execute()`.

 usageText - starts with app name, that will be referenced later. Otherwise the text will be printed out to the user.

 desc - information text for the end user.

 If you want to inspect all built configurations, then call `Initialize()`

 For examples look into `tests/pungi_test.go` or `testapp/testapp.go` or `testappMulti/testapp_multi.go`.
*/
func New(usageText, desc string) *pungiBuilder {
	return &pungiBuilder{
		usageText:         usageText,
		desc:              desc,
		defaultConfigFile: "config.toml",
		commands:          make(map[string]*Command),
		keys:              make(map[string]*key),
		confs:             make(map[string]*Conf),
	}
}

// Initializes configuration only from a config file. Useful for using inside tests.
func NewConfigFileOnly(appName, filePath string) (*Pungi, error) {
	viper.SetConfigFile(filePath)
	err := viper.ReadInConfig()
	if err == nil {
		println("Using config file: " + viper.ConfigFileUsed())
		return &Pungi{
			appName:        appName,
			configFileUsed: viper.ConfigFileUsed(),
		}, nil
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Could not load config from %s\n Error: %v", viper.ConfigFileUsed(), err)
		return nil, err
	}
}

// Initializes and executes Pungi. In case of errors panics.
func (p *pungiBuilder) Execute() {
	if pungi, err := p.Initialize(); err != nil {
		panic(err)
	} else {
		if err := pungi.Execute(); err != nil {
			panic(err)
		}
	}
}

// Initializes and returns the Pungi.
// Does not execute the runnables.
func (p *pungiBuilder) Initialize() (*Pungi, error) {
	if err := p.validateInput(); err != nil {
		return nil, err
	}

	pungi := &Pungi{}
	p.appName = firstWord(p.usageText)
	p.initRootCommand(pungi)
	if p.runnable != nil {
		p.initRootKeys()
	}

	if len(p.commands) > 0 {
		for _, cmd := range p.commands {
			p.initSubCommand(cmd)
		}
	}

	pungi.appName = p.appName
	pungi.confs = p.confs
	pungi.rootCmd = p.rootCommand

	// Adds "normal" flags too. I.e. glog
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	// Hack to convince flag that flags are parsed.
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		return nil, err
	}

	return pungi, nil
}

func (p *pungiBuilder) DefaultConfigFile(filename string) *pungiBuilder {
	p.defaultConfigFile = filename
	return p
}

func (p *pungiBuilder) Key(name string, value interface{}, desc string) *pungiBuilder {
	p.keys[name] = &key{
		name:  name,
		value: value,
		desc:  desc,
	}
	return p
}

// Defines a subcommand. Construct commands using `Pungi.Cmd()` function.
func (p *pungiBuilder) Cmd(command *Command) *pungiBuilder {
	p.commands[command.cmdName] = command
	return p
}

// Defines runnable function. When it's defined, no subcommands may be defined.
func (p *pungiBuilder) Run(runnable func(conf *Conf, args []string) error) *pungiBuilder {
	p.runnable = runnable
	return p
}

// Defines validation rules for arguments. For pre-defined validation rules see `cobra/args.go` file.
func (p *pungiBuilder) Args(args cobra.PositionalArgs) *pungiBuilder {
	p.args = args
	return p
}

func merge(a, b map[string]*key) (out map[string]*key) {
	out = make(map[string]*key)
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return
}

func (p *pungiBuilder) validateInput() error {
	if err := p.validateKeys(); err != nil {
		return err
	}
	if p.runnable != nil && len(p.commands) > 0 && p.args == nil {
		return errors.New("If you define main and subcommands, then you need to define arguments for the main command.")
	}
	return nil
}

func (p *pungiBuilder) initRootKeys() {
	for _, key := range p.keys {
		initFlag(p.rootCommand, key)
		p.bindRootKey(p.rootCommand, key.name)
	}
}

func (p *pungiBuilder) initSubCommand(cmd *Command) {

	var conf = &Conf{appName: p.appName, cmdName: cmd.cmdName}
	runnable := cmd.runnable
	p.confs[cmd.cmdName] = conf

	cobraRun := func(cobraCmd *cobra.Command, args []string) error {
		return runnable(conf, args)
	}
	cobraCmd := &cobra.Command{
		Use:   cmd.usageText,
		Short: cmd.desc,
		Args:  cmd.args,
		RunE:  cobraRun,
	}

	p.rootCommand.AddCommand(cobraCmd)

	allKeys := merge(p.keys, cmd.keys)
	for _, key := range allKeys {
		initFlag(cobraCmd, key)
		p.bindSubCmdKey(cobraCmd, cmd.cmdName, key.name)
	}
}

func (p *pungiBuilder) initRootCommand(pungi *Pungi) {
	var cfgFile string
	cobra.OnInitialize(func() {
		if err := p.initViper(pungi, cfgFile); err != nil {
			panic(err)
		}
	})

	p.confs[rootKey] = newRootConf(p.appName)
	var rootRunnable func(cmd *cobra.Command, args []string) error

	if p.runnable != nil {
		rootRunnable = func(cobraCmd *cobra.Command, args []string) error {
			return p.runnable(p.confs[rootKey], args)
		}
	}

	p.rootCommand = &cobra.Command{
		Use:   p.usageText,
		Short: p.desc,
		Args:  p.args,
		RunE:  rootRunnable,
	}

	p.rootCommand.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.toml)")
}

func (p *pungiBuilder) bindSubCmdKey(command *cobra.Command, cmdName, key string) {
	confKey := formatCommandConfKey(p.appName, cmdName, key)
	envKey := formatCommandEnvKey(p.appName, cmdName, key)

	if err := viper.BindPFlag(confKey, command.Flags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(confKey, envKey); err != nil {
		panic(err)
	}
}

func (p *pungiBuilder) bindRootKey(command *cobra.Command, key string) {
	confKey := formatRootConfKey(p.appName, key)
	envKey := formatRootEnvKey(p.appName, key)

	if err := viper.BindPFlag(confKey, command.Flags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(confKey, envKey); err != nil {
		panic(err)
	}
}

func (p *pungiBuilder) validateKeys() error {
	allKeys := p.keys
	for _, cmd := range p.commands {
		allKeys = merge(allKeys, cmd.keys)
	}
	for _, key := range allKeys {
		switch key.value.(type) {
		case string:
		case int:
		case bool:
		case float64:
		default:
			return errors.New("Not suppored value type: " + reflect.TypeOf(key.value).String())
		}
	}
	return nil
}

// Panics because keys are validated before
func initFlag(command *cobra.Command, key *key) {
	switch v := key.value.(type) {
	case string:
		command.Flags().String(key.name, v, key.desc)
	case int:
		command.Flags().Int(key.name, v, key.desc)
	case bool:
		command.Flags().Bool(key.name, v, key.desc)
	case float64:
		command.Flags().Float64(key.name, v, key.desc)
	default:
		panic("Unknown value type: " + reflect.TypeOf(key.value).String())
	}
}

func (p *pungiBuilder) initViper(pungi *Pungi, cfgFileFlag string) error {
	if cfgFileFlag != "" {
		viper.SetConfigFile(cfgFileFlag)
	} else {
		if cfgFileFromEnv := os.Getenv(formatRootEnvKey(p.appName, "CONFIG")); cfgFileFromEnv != "" {
			viper.SetConfigFile(cfgFileFromEnv)
		} else {
			viper.SetConfigFile(p.defaultConfigFile)
		}
	}
	viper.SetEnvPrefix(p.appName)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err == nil {
		println("Using config file: " + viper.ConfigFileUsed())
	} else {
		if viper.ConfigFileUsed() == p.defaultConfigFile {
			println("Default config file not found")
			err = nil
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Could not load config from %s\n Error: %v", viper.ConfigFileUsed(), err)
		}
	}
	pungi.configFileUsed = viper.ConfigFileUsed()
	return err
}

type pungiBuilder struct {
	appName, usageText, desc string
	keys                     map[string]*key
	commands                 map[string]*Command
	runnable                 Runnable
	confs                    map[string]*Conf
	rootCommand              *cobra.Command
	defaultConfigFile        string
	args                     cobra.PositionalArgs
}

type Runnable = func(conf *Conf, args []string) error

type key struct {
	name, desc string
	value      interface{}
}

// Pungi contains all computed configurations.
// In most cases this is not needed.
type Pungi struct {
	confs          map[string]*Conf
	rootCmd        *cobra.Command
	configFileUsed string
	appName        string
}

// Returns the root config. The values that are shared by commands
// If no commands are defined, then the application config.
func (p *Pungi) RootConfig() *Conf {
	if p.confs == nil {
		p.confs = make(map[string]*Conf)
	}
	if _, ok := p.confs[rootKey]; !ok {
		p.confs[rootKey] = newRootConf(p.appName)
	}
	return p.confs[rootKey]
}

func newRootConf(appName string) *Conf {
	return &Conf{appName: appName}
}

// Returns specific command configuration
func (p *Pungi) Config(cmdName string) *Conf {
	if conf, ok := p.confs[cmdName]; ok {
		return conf
	}
	return &Conf{
		appName: p.appName,
		cmdName: cmdName,
	}
}

// Returns the config file name that is used
func (p *Pungi) ConfigFileUsed() string {
	return p.configFileUsed
}

// args - arguments for executable. By default: `os.Args[1:]`
func (p *Pungi) Execute(args ...string) error {
	if len(args) > 0 {
		p.rootCmd.SetArgs(args)
	}
	return p.rootCmd.Execute()
}
