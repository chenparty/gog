package zwriter

import (
	"github.com/nats-io/nats.go"
)

type NATSWriterOption struct {

	// NATS client
	Connection *nats.Conn
	Subject    string
}

// NewNATSWriter 创建一个NATS写入器
func (o NATSWriterOption) NewNATSWriter() *NATSWriter {
	if o.Connection == nil {
		panic("missing NATS connection")
	}

	if o.Subject == "" {
		panic("missing NATS subject")
	}

	return &NATSWriter{
		option: o,
	}
}

type NATSWriter struct {
	option NATSWriterOption
}

func (w *NATSWriter) Write(p []byte) (n int, err error) {
	err = w.option.Connection.Publish(w.option.Subject, p)
	if err == nil {
		n = len(p)
	}
	return
}
