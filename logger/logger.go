// Copyright 2019 Samaritan Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"log/syslog"
	"os"

	"github.com/tevino/log"
)

const (
	defaultFlags    = log.LstdFlags | log.Lshortfile | log.Lmicroseconds
	defaultPriority = syslog.LOG_INFO | syslog.LOG_USER
)

var (
	logger log.Logger
)

func init() {
	logger = log.NewLogger(os.Stdout, defaultFlags)
	logger.SetCallerOffset(1)
	logger.SetOutputLevel(log.INFO)
}

func InitSyslog(target, tag string) {
	writer, err := newSysLogWriter(target, defaultPriority, tag)
	if err != nil {
		logger.Fatal("Init syslog fail: ", err)
	}
	logger = log.NewLogger(writer, defaultFlags)
	logger.SetCallerOffset(1)
}

func SetLevel(level string) {
	logger.SetOutputLevel(log.LevelFromString(level))
	logger.Info("Log level: ", logger.OutputLevel())
}

// Get gets the logger.
func Get() log.Logger {
	return logger
}

// Debug calls the same method on global logger.
func Debug(a ...interface{}) {
	logger.Debug(a...)
}

// Debugf calls the same method on global logger.
func Debugf(f string, a ...interface{}) {
	logger.Debugf(f, a...)
}

// Info calls the same method on global logger.
func Info(a ...interface{}) {
	logger.Info(a...)
}

// Infof calls the same method on global logger.
func Infof(f string, a ...interface{}) {
	logger.Infof(f, a...)
}

// Warn calls the same method on global logger.
func Warn(a ...interface{}) {
	logger.Warn(a...)
}

// Warnf calls the same method on global logger.
func Warnf(f string, a ...interface{}) {
	logger.Warnf(f, a...)
}

// Fatal calls the same method on global logger.
func Fatal(a ...interface{}) {
	logger.Fatal(a...)
}

// Fatalf calls the same method on global logger.
func Fatalf(f string, a ...interface{}) {
	logger.Fatalf(f, a...)
}
