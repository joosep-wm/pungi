package pungi

import (
	"fmt"

	"strings"

	"github.com/spf13/viper"
)

// Conf is configuration that's related to one specific command
type Conf struct {
	cmdName string
	appName string
}

func (c *Conf) fullKey(key string) string {
	if c.cmdName == "" {
		return formatRootConfKey(c.appName, key)
	}
	return formatCommandConfKey(c.appName, c.cmdName, key)
}

func (c *Conf) GetBool(key string) bool {
	return viper.GetViper().GetBool(c.fullKey(key))
}
func (c *Conf) GetInt(key string) int {
	return viper.GetViper().GetInt(c.fullKey(key))
}
func (c *Conf) GetFloat64(key string) float64 {
	return viper.GetViper().GetFloat64(c.fullKey(key))
}
func (c *Conf) GetString(key string) string {
	return viper.GetViper().GetString(c.fullKey(key))
}
func (c *Conf) Set(key string, value interface{}) {
	viper.GetViper().Set(c.fullKey(key), value)
}

// Low level constructor, useful for tests.
func NewConf(appName, cmdName string) *Conf {
	return &Conf{
		appName: appName,
		cmdName: cmdName,
	}
}

// Returns all values defined in this configuration instance
func (c *Conf) AllValues() map[string]interface{} {
	all := viper.GetViper().AllSettings()
	if app, ok := all[c.appName].(map[string]interface{}); ok {
		if c.cmdName == "" {
			return app
		} else {
			if cmd, ok := app[c.cmdName].(map[string]interface{}); ok {
				return cmd
			} else {
				return make(map[string]interface{})
			}
		}
	} else {
		return make(map[string]interface{})
	}
}

func formatRootConfKey(appName, key string) string {
	return strings.ToLower(fmt.Sprintf("%s.%s", appName, key))
}
func formatCommandConfKey(appName, cmdName, key string) string {
	return strings.ToLower(fmt.Sprintf("%s.%s.%s", appName, cmdName, key))
}
func formatRootEnvKey(appName, key string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", appName, key))
}
func formatCommandEnvKey(appName, cmdName, key string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s_%s", appName, cmdName, key))
}
