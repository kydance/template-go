package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kydenul/log"
	"github.com/spf13/viper"
)

var Logger *log.ZiwiLog

const (
	defaultConfigDir  = "etc"
	defaultConfigName = "ziwi"

	envPrefix = "ZIWI"
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user home directory: %v", err))
	}

	viper.AddConfigPath(filepath.Join(home, defaultConfigDir)) // $HOME/defaultConfigDir
	viper.AddConfigPath(filepath.Join(".", defaultConfigDir))  // ./defaultConfigDir
	viper.AddConfigPath(filepath.Join("/Users/kyden/git-space/mcp-svr", defaultConfigDir))

	viper.SetConfigType("yaml")
	viper.SetConfigName(defaultConfigName)

	// Read matched environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file.
	// If a config file is specified, use it. Otherwise, search in defaultConfigDir.
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Failed to read viper config file: %v", err))
	}

	Logger = log.NewLogger(log.NewOptions().
		WithPrefix(viper.GetString("log.prefix")).
		WithDirectory(viper.GetString("log.directory")).
		WithLevel(viper.GetString("log.level")).
		WithTimeLayout(viper.GetString("log.time-layout")).
		WithFormat(viper.GetString("log.format")).
		WithDisableCaller(viper.GetBool("log.disable-caller")).
		WithDisableStacktrace(viper.GetBool("log.disable-stacktrace")).
		WithDisableSplitError(viper.GetBool("log.disable-split-error")).
		WithMaxSize(viper.GetInt("log.max-size")).
		WithMaxBackups(viper.GetInt("log.max-backups")).
		WithCompress(viper.GetBool("log.compress")))
}

func main() {
	Logger.Infoln("Starting greeting server...")
}
