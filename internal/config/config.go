package config

import (
	"fmt"

	splunk "github.com/pvik/go-splunk-rest"
	"gorm.io/gorm"

	log "log/slog"

	"github.com/BurntSushi/toml"
)

// DBConfig holds connection details to a Database
type DBConfig struct {
	DBType       string   `toml:"db-type"` // postgres , sqlserver , mysql , sqlite
	Host         string   `toml:"host"`
	Port         int      `toml:"port"`
	Username     string   `toml:"username"`
	Password     string   `toml:"password"`
	DBName       string   `toml:"db-name"`
	ExtraParam   string   `toml:"extra-connection-parameters"`
	DBConnection *gorm.DB `toml:"-"`
}

// LogConfig holds log information
type LogConfig struct {
	Format string `toml:"format"`
	Output string `toml:"output"`
	Dir    string `toml:"log-directory"`
	Level  string `toml:"level"`
}

type SplunkFieldTranslation struct {
	DBColumnName     string            `toml:"db-column-name"`
	ValueTranslation map[string]string `toml:"value-translations"`
}

type QueryDetails struct {
	SplunkName      string                            `toml:"from-splunk-name"`
	DBName          string                            `toml:"to-db-name"`
	Search          string                            `toml:"search-query"`
	AllowPartition  bool                              `toml:"allow-partition"`
	StartTime       string                            `toml:"start-time"`
	EndTime         string                            `toml:"end-time"`
	DBTableName     string                            `toml:"db-table"`
	DBPrimaryKeyCol string                            `toml:"db-table-primary-key-column"`
	IncludeFields   []string                          `toml:"include-fields"`
	SplunkFields    map[string]SplunkFieldTranslation `toml:"splunk-fields-translation"`
}

// Config holds all the details from config.toml passed to application
type Config struct {
	Log     LogConfig                    `toml:"log"`
	DB      map[string]DBConfig          `toml:"database"`
	Splunk  map[string]splunk.Connection `toml:"splunk"`
	Queries map[string]QueryDetails      `toml:"queries"`
}

// AppConf package global has values parsed from config.toml
var AppConf Config

// InitConfig Initializes AppConf
// It reads in the Config file at configPath and populates AppConf
func InitConfig(configPath string) {
	// log.Info("Reading in Config File",
	// 	"file", configPath)

	if _, err := toml.DecodeFile(configPath, &AppConf); err != nil {
		log.Error("unable to parse config toml file",
			"error", err)
		panic(fmt.Errorf("unable to parse config toml file"))
	}
}
