package log

import (
	"bufio"
	"os"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var logger *Logger

type Logger struct {
	*logrus.Logger
	Pid int
}

func init() {
	logger = NewLog(Config{Console: true})
}

// Config 日志配置
type Config struct {
	Console         bool   // 是否把日志打印到控制台 默认: false
	StorageLocation string // 文件存储路径 默认: ../logs/
	Level           string // 日志等级 panic fatal error warn info debug trace 默认: debug
	ModuleName      string // 模型名称
	RotationTime    int    // 日志旋转时间(小时) 默认: 24
	MaxRemainNum    uint   // 最大保留文件数 默认: 2
}

// NewPrivateLog 自定义全局日志
func NewPrivateLog(cfg Config) {
	logger = NewLog(cfg)
}

func NewLog(cfg Config) *Logger {
	if cfg.StorageLocation == "" {
		cfg.StorageLocation = "../logs/"
	}
	if cfg.Level == "" {
		cfg.Level = "error"
	}
	if cfg.RotationTime <= 0 {
		cfg.RotationTime = 24
	}
	if cfg.MaxRemainNum == 0 {
		cfg.MaxRemainNum = 2
	}
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.DebugLevel
	}

	var newLogger = logrus.New()
	newLogger.SetLevel(level)

	// Close std console output
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err.Error())
	}
	writer := bufio.NewWriter(src)
	newLogger.SetOutput(writer)
	if cfg.Console {
		newLogger.SetOutput(os.Stdout)
	}

	// 日志控制台打印样式设置
	newLogger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		HideKeys:        false,
		FieldsOrder:     []string{"PID", "FilePath", "OperationID"},
	})
	// 文件名和行号显示钩子
	newLogger.AddHook(newFileHook())

	// 日志文件分割
	hook := NewLfsHook(cfg.StorageLocation, time.Duration(cfg.RotationTime)*time.Hour, cfg.MaxRemainNum, cfg.ModuleName)
	newLogger.AddHook(hook)
	return &Logger{
		newLogger,
		os.Getpid(),
	}
}

func NewLfsHook(storageLocation string, rotationTime time.Duration, maxRemainNum uint, moduleName string) logrus.Hook {
	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: initRotateLogs(storageLocation, rotationTime, maxRemainNum, "all", moduleName),
		logrus.InfoLevel:  initRotateLogs(storageLocation, rotationTime, maxRemainNum, "all", moduleName),
		logrus.WarnLevel:  initRotateLogs(storageLocation, rotationTime, maxRemainNum, "all", moduleName),
		logrus.ErrorLevel: initRotateLogs(storageLocation, rotationTime, maxRemainNum, "all", moduleName),
	}, &nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		HideKeys:        false,
		FieldsOrder:     []string{"PID", "FilePath", "OperationID"},
	})
	return lfsHook
}

func initRotateLogs(storageLocation string, rotationTime time.Duration, maxRemainNum uint, level string, moduleName string) *rotatelogs.RotateLogs {
	if moduleName != "" {
		moduleName = moduleName + "."
	}
	writer, err := rotatelogs.New(
		storageLocation+moduleName+level+"."+"%Y-%m-%d",
		rotatelogs.WithRotationTime(rotationTime),
		rotatelogs.WithRotationCount(maxRemainNum),
	)
	if err != nil {
		panic(err.Error())
	} else {
		return writer
	}
}

func Println(args ...interface{}) {
	logger.WithFields(logrus.Fields{
		"PID": logger.Pid,
	}).Println(args)
}

func Info(args ...interface{}) {
	logger.WithFields(logrus.Fields{
		"PID": logger.Pid,
	}).Infoln(args)
}

func Error(args ...interface{}) {
	logger.WithFields(logrus.Fields{
		"PID": logger.Pid,
	}).Errorln(args)
}

func Warn(args ...interface{}) {
	logger.WithFields(logrus.Fields{
		"PID": logger.Pid,
	}).Warnln(args)
}

func Debug(args ...interface{}) {
	logger.WithFields(logrus.Fields{
		"PID": logger.Pid,
	}).Debugln(args)
}

func Printf(format string, args ...interface{}) {
	logger.WithFields(logrus.Fields{
		"PID": logger.Pid,
	}).Printf(format, args)
}

// Infof logs an info message
func Infof(format string, args ...interface{}) {
	logger.WithFields(logrus.Fields{}).Infof(format, args)
}

// Debugf logs a debug message.
func Debugf(format string, args ...interface{}) {
	logger.WithFields(logrus.Fields{}).Debugf(format, args)
}

// Warnf logs an error message.
func Warnf(format string, args ...interface{}) {
	logger.WithFields(logrus.Fields{}).Errorf(format, args...)
}

//Errorf logs an error message.
func Errorf(format string, args ...interface{}) {
	logger.WithFields(logrus.Fields{}).Errorf(format, args...)
}
