package dbo

import "gorm.io/gorm/logger"

type LogLevel string

const (
	Silent LogLevel = "Silent"
	Error  LogLevel = "Error"
	Warn   LogLevel = "Warn"
	Info   LogLevel = "Info"
)

func (l LogLevel) String() string {
	return string(l)
}

func (l LogLevel) GormLogLevel() logger.LogLevel {
	switch l {
	case Silent:
		return logger.Silent
	case Error:
		return logger.Error
	case Warn:
		return logger.Warn
	case Info:
		return logger.Info
	default:
		return logger.Silent
	}
}
