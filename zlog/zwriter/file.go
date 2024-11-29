package zwriter

import (
	"gopkg.in/natefinch/lumberjack.v2"
)

type FileWriterOption struct {
	FileName string
	MaxSize  int // MB
	MaxAge   int // DAYS
	Compress bool
}

func (o FileWriterOption) NewFileWriter() *lumberjack.Logger {
	if o.FileName == "" {
		o.FileName = "log/app.log"
	}

	if o.MaxSize == 0 {
		o.MaxSize = 10
	}

	if o.MaxAge == 0 {
		o.MaxAge = 30
	}

	return &lumberjack.Logger{
		Filename:  o.FileName,
		MaxSize:   o.MaxSize, // MB
		MaxAge:    o.MaxAge,  //days
		Compress:  o.Compress,
		LocalTime: true,
	}
}
