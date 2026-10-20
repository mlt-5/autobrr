// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/internal/logger"
	"github.com/autobrr/autobrr/pkg/errors"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var EnvVarPrefix = "AUTOBRR__"

var configTemplate = `# config.toml

# autobrr logs file
# If not defined, logs to stdout
# Make sure to use forward slashes and include the filename with extension. eg: "/path/to/logs/autobrr.log", "C:/autobrr/logs/autobrr.log"
#
# Optional
#
#logPath = "/path/to/logs/autobrr.log"

# Log level
#
# Default: "INFO"
#
# Options: "ERROR", "WARN", "INFO", "DEBUG", "TRACE"
#
logLevel = "INFO"

# Log Max Size
#
# Default: 10
#
# Max log size in megabytes
#
#logMaxSize = 10

# Log Max Backups
#
# Default: 5
#
# Max amount of old log files
#
#logMaxBackups = 5

# Database Max Backups
#
# Default: 5
#
#databaseMaxBackups = 5

# Custom definitions
#
#customDefinitions = "test/definitions"
`

func (c *AppConfig) writeConfig(configPath string, configFile string) error {
	cfgPath := filepath.Join(configPath, configFile)

	// check if configPath exists, if not create it
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	// check if config exists, if not create it
	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(cfgPath)
		if err != nil { // perm 0666
			// handle failed create
			log.Printf("error creating file: %q", err)
			return err
		}
		defer f.Close()

		if _, err = f.WriteString(configTemplate); err != nil {
			log.Printf("error writing contents to file: %v %q", configPath, err)
			return err
		}

		return f.Sync()
	}

	return nil
}

type AppConfig struct {
	Config *domain.Config
	m      *sync.Mutex
}

func New(configPath string) *AppConfig {
	c := &AppConfig{
		m: new(sync.Mutex),
	}
	c.defaults()
	c.Config.ConfigPath = configPath

	c.load(configPath)
	c.loadFromEnv()

	return c
}

func (c *AppConfig) defaults() {
	c.Config = &domain.Config{
		LogLevel:            "INFO",
		LogPath:             "",
		LogMaxSize:          10,
		LogMaxBackups:       5,
		CustomDefinitions:   "",
		DatabaseAutoMigrate: true,
		DatabaseMaxBackups:  5,
		DatabaseDSN:         "",
	}
}

func (c *AppConfig) loadFromEnv() {
	if v := GetEnvStr("CUSTOM_DEFINITIONS"); v != "" {
		c.Config.CustomDefinitions = v
	}

	// Logs configuration
	if v := GetEnvStr("LOG_LEVEL"); v != "" {
		c.Config.LogLevel = v
	}

	if v := GetEnvStr("LOG_PATH"); v != "" {
		c.Config.LogPath = v
	}

	if v := GetEnvInt("LOG_MAX_SIZE"); v > 0 {
		c.Config.LogMaxSize = v
	}

	if v := GetEnvInt("LOG_MAX_BACKUPS"); v > 0 {
		c.Config.LogMaxBackups = v
	}

	// Database configuration
	if v := GetEnvStr("DATABASE_DSN"); v != "" {
		c.Config.DatabaseDSN = v
	}

	if v := GetEnvInt("DATABASE_MAX_BACKUPS"); v > 0 {
		c.Config.DatabaseMaxBackups = v
	}
}

func GetEnvStr(key string) string {
	// first check if we have a variable with a _FILE ending
	// commonly used for docker secrets and similar
	if filePath := os.Getenv(EnvVarPrefix + key + "_FILE"); filePath != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Could not read file: %s err: %q", filePath, err)
			return ""
		}
		return strings.TrimSpace(string(content))
	}

	if v := os.Getenv(EnvVarPrefix + key); v != "" {
		return v
	}

	return ""
}

func GetEnvInt(key string) int {
	value := GetEnvStr(key)

	i, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0
	}
	return int(i)
}

func (c *AppConfig) load(configPath string) {
	viper.SetConfigType("toml")

	// clean trailing slash from configPath
	configPath = path.Clean(configPath)

	if configPath != "" {
		//viper.SetConfigName("config")

		// check if path and file exists
		// if not, create path and file
		if err := c.writeConfig(configPath, "config.toml"); err != nil {
			log.Printf("write error: %q", err)
		}

		viper.SetConfigFile(path.Join(configPath, "config.toml"))
	} else {
		viper.SetConfigName("config")

		// Search config in directories
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/autobrr")
		viper.AddConfigPath("$HOME/.autobrr")
	}

	// read config
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("config read error: %q", err)
	}

	if err := viper.Unmarshal(c.Config); err != nil {
		log.Fatalf("Could not unmarshal config file: %v: err %q", viper.ConfigFileUsed(), err)
	}
}

func (c *AppConfig) DynamicReload(log zerolog.Logger) {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		c.m.Lock()
		defer c.m.Unlock()

		logLevel := viper.GetString("logLevel")
		c.Config.LogLevel = logLevel
		logger.SetLevel(c.Config.LogLevel)

		logPath := viper.GetString("logPath")
		c.Config.LogPath = logPath

		log.Debug().Msg("config file reloaded!")
	})
}

func (c *AppConfig) UpdateConfig() error {
	filePath := path.Join(c.Config.ConfigPath, "config.toml")

	f, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "could not read config file: %s", filePath)
	}

	lines := strings.Split(string(f), "\n")
	lines = c.processLines(lines)

	output := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		return errors.Wrap(err, "could not write config file: %s", filePath)
	}

	return nil
}

func (c *AppConfig) processLines(lines []string) []string {
	// keep track of not found values to append at bottom
	var (
		foundLineLogLevel = false
		foundLineLogPath  = false
	)

	for i, line := range lines {
		if !foundLineLogLevel && strings.Contains(line, "logLevel =") {
			lines[i] = fmt.Sprintf(`logLevel = "%s"`, c.Config.LogLevel)
			foundLineLogLevel = true
		}
		if !foundLineLogPath && strings.Contains(line, "logPath =") {
			if c.Config.LogPath == "" {
				// Check if the line already has a value
				matches := strings.Split(line, "=")
				if len(matches) > 1 && strings.TrimSpace(matches[1]) != `""` {
					lines[i] = line // Preserve the existing line
				} else {
					lines[i] = `#logPath = ""`
				}
			} else {
				lines[i] = fmt.Sprintf("logPath = \"%s\"", c.Config.LogPath)
			}
			foundLineLogPath = true
		}
	}

	// append missing vars to bottom
	if !foundLineLogLevel {
		lines = append(lines, "# Log level")
		lines = append(lines, "#")
		lines = append(lines, `# Default: "INFO"`)
		lines = append(lines, "#")
		lines = append(lines, `# Options: "ERROR", "WARN", "INFO", "DEBUG", "TRACE"`)
		lines = append(lines, "#")
		lines = append(lines, fmt.Sprintf(`logLevel = "%s"`, c.Config.LogLevel))
	}

	if !foundLineLogPath {
		lines = append(lines, "# Log Path")
		lines = append(lines, "#")
		lines = append(lines, "# Optional")
		lines = append(lines, "#")
		if c.Config.LogPath == "" {
			lines = append(lines, `#logPath = ""`)
		} else {
			lines = append(lines, fmt.Sprintf(`logPath = "%s"`, c.Config.LogPath))
		}
	}

	return lines
}
