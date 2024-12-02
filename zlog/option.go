package zlog

import (
	"github.com/chenparty/gog/zlog/zwriter"
	"github.com/nats-io/nats.go"
	"io"
	"os"
)

type Options struct {
	Mode  LogMode
	Level Level

	FileWriterOption zwriter.FileWriterOption
	NATSWriterOption zwriter.NATSWriterOption
}

type Option func(*Options)

func (o Options) newWriter() io.Writer {
	switch o.Mode {
	case FILE:
		return o.FileWriterOption.NewFileWriter()
	case NATS:
		return o.NATSWriterOption.NewNATSWriter()
	default:
		return os.Stdout
	}
}

func FileAttr(name string, maxSize int, maxAge int, compress bool) Option {
	return func(o *Options) {
		o.FileWriterOption = zwriter.FileWriterOption{
			FileName: name,
			MaxSize:  maxSize,
			MaxAge:   maxAge,
			Compress: compress,
		}
	}
}

func NATSAttr(conn *nats.Conn, subject string) Option {
	return func(o *Options) {
		o.NATSWriterOption = zwriter.NATSWriterOption{
			Connection: conn,
			Subject:    subject,
		}
	}
}
