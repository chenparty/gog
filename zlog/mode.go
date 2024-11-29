package zlog

import (
	"errors"
	"fmt"
	"strings"
)

type LogMode int

const (
	STDOUT LogMode = iota
	FILE
	NATS
)

func (m *LogMode) parse(s string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("zlog: log mode string %q: %w", s, err)
		}
	}()
	switch strings.ToUpper(s) {
	case "STDOUT":
		*m = STDOUT
	case "FILE":
		*m = FILE
	case "NATS":
		*m = NATS
	default:
		err = errors.New("unknown name")
	}
	return
}
