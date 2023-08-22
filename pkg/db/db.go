package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	log "log/slog"
)

// Init initializes a database connection and sets up package global DB var
// to be usable through the rest of the application
func Connect(dbtype, host string, port int, dbname, user, password, extraParamStr, logLevelStr string) (*gorm.DB, error) {
	log.Debug("init db", "db-type", dbtype, "host", host, "db-name", dbname)
	var err error

	// var extraParamStr string
	// if !sslmode {
	// 	extraParamStr = " sslmode=disable"
	// }

	dbLogLevel := logger.Warn
	if logLevelStr == "info" {
		dbLogLevel = logger.Info
	}

	var DB *gorm.DB
	gormConf := &gorm.Config{
		Logger: &GormLogger{
			LogLevel:                  dbLogLevel,
			LogTrace:                  false,
			SlowThreshold:             5 * time.Second,
			IgnoreRecordNotFoundError: false, // Ignore ErrRecordNotFound error for logger
		},
	}

	if dbtype == "postgres" {
		connString := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s%s",
			host, port, user, dbname, password, extraParamStr)
		DB, err = gorm.Open(postgres.Open(connString),
			gormConf)
	} else if dbtype == "sqlserver" {
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s%s",
			user, password, host, port, dbname, extraParamStr)
		DB, err = gorm.Open(sqlserver.Open(dsn), gormConf)
	} else if dbtype == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local%s",
			user, password, host, port, dbname, extraParamStr)
		DB, err = gorm.Open(mysql.Open(dsn), gormConf)
	} else if dbtype == "sqlite" {
		DB, err = gorm.Open(sqlite.Open(dbname), gormConf)
	}

	if err != nil {
		log.Error("Unable to open database",
			"db-type", dbtype,
			"host", host,
			"port", port,
			"db-name", dbname,
			"user", user,
			"error", err,
		)
		return DB, err
	}

	return DB, nil
}
