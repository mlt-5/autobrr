// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

type Config struct {
	ConfigPath          string
	Host                string `toml:"host"`
	Port                int    `toml:"port"`
	CorsAllowedOrigins  string `toml:"corsAllowedOrigins"`
	LogLevel            string `toml:"logLevel"`
	LogPath             string `toml:"logPath"`
	LogMaxSize          int    `toml:"logMaxSize"`
	LogMaxBackups       int    `toml:"logMaxBackups"`
	BaseURL             string `toml:"baseUrl"`
	BaseURLModeLegacy   bool   `toml:"baseUrlModeLegacy"`
	CustomDefinitions   string `toml:"customDefinitions"`
	DatabaseDSN         string `toml:"databaseDSN"`
	DatabaseMaxBackups  int    `toml:"databaseMaxBackups"`
	DatabaseAutoMigrate bool   `toml:"databaseAutoMigrate"`
}

type ConfigUpdate struct {
	Host     *string `json:"host,omitempty"`
	Port     *int    `json:"port,omitempty"`
	LogLevel *string `json:"log_level,omitempty"`
	LogPath  *string `json:"log_path,omitempty"`
	BaseURL  *string `json:"base_url,omitempty"`
}
