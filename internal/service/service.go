package service

import (
	// "encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	c "github.com/pvik/splunk2db/internal/config"

	log "log/slog"
)

const sslVerify = false

var (
	logFileHandle *os.File
)

// InitService initialize the microservice
// It does the following:
//   - initialize config file (passed in as command line arg)
//   - Setup Logging
func InitService() {
	_, serviceName := filepath.Split(os.Args[0])

	commit := getCommit()

	log.Info(serviceName+" starting", "version", "0.2.1", "git commit", commit)

	var confFile string
	flag.StringVar(&confFile, "conf", "", "config file for microservice")
	flag.Parse()

	if confFile == "" {
		log.Error("Please provide config file as command line arg")
		flag.PrintDefaults()
		os.Exit(10)
	}

	c.InitConfig(confFile)

	var logger *log.Logger
	var logOuput io.Writer
	var logLevel log.Leveler
	// Initialize Logfile
	if c.AppConf.Log.Output == "file" {
		var err error
		logFile := path.Join(c.AppConf.Log.Dir,
			fmt.Sprintf("%s.log",
				serviceName))
		logFileHandle, err = os.OpenFile(logFile,
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Error("unable to open file")
			panic("unable to open logfile")
		}
		log.Info("switching log output to file", "file", logFile)
		logOuput = logFileHandle
	} else {
		logOuput = os.Stdout
	}

	// set log level
	switch strings.ToLower(c.AppConf.Log.Level) {
	case "debug":
		logLevel = log.LevelDebug
	case "info":
		logLevel = log.LevelInfo
	case "warn":
		logLevel = log.LevelWarn
	case "error":
		logLevel = log.LevelError
	}

	if c.AppConf.Log.Format == "json" {
		logger = log.New(log.NewJSONHandler(logOuput,
			&log.HandlerOptions{
				// AddSource: true,
				Level: logLevel,
				ReplaceAttr: func(groups []string, a log.Attr) log.Attr {
					if a.Key == "time" {
						a.Value = log.StringValue(a.Value.Time().Format(time.DateTime))
					}
					return a
				},
			}))
	} else {
		// Default to TextFormatter
		logger = log.New(log.NewTextHandler(logOuput,
			&log.HandlerOptions{
				// AddSource: true,
				Level: logLevel,
				ReplaceAttr: func(groups []string, a log.Attr) log.Attr {
					if a.Key == "time" {
						a.Value = log.StringValue(a.Value.Time().Format(time.DateTime))
					}
					return a
				},
			}))
	}

	log.SetDefault(logger)

	log.Info(serviceName + " service initialized")
}

// Shutdown closes any open files or pipes the microservice started
// It does the following:
//   - Disconnect to DB
func Shutdown() {
	log.Debug("shutdown service")
	//db.Close()

	if logFileHandle != nil {
		log.Debug("terminate log file handle")
		// Revert logging back to StdOut
		//log.SetOutput(os.Stdout)
		logFileHandle.Close()
	}
}

func getCommit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}
