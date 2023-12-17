// Summary: Configuration for the logger
// Description: This file is used to configure the logger for the whole application.
// Author: Egor Pristavka
// Date: 2023-03-27

package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var (
	Logger *logrus.Logger
)

func init() {
	// Setup logger for the whole application, add the hook to it, and configure it to log as JSON instead of the default ASCII formatter.
	Logger = &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			PrettyPrint:     true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.DebugLevel,
	}
}
