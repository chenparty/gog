package zlog

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const ctxTraceIDKey = "trace_id"

// TraceHook 是一个自定义 Hook，用于为每个日志条目添加 trace_id
type TraceHook struct{}

// Run 实现 zerolog.Hook 接口，允许修改日志条目
func (h *TraceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	// 将 trace_id 添加到日志条目
	traceID, _ := e.GetCtx().Value(ctxTraceIDKey).(string)
	e.Str("trace_id", traceID)
}

func ContextWithValue(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxTraceIDKey, traceID)
}

func NewTraceContext(ctx context.Context) context.Context {
	traceID := uuid.New().String()
	return ContextWithValue(ctx, traceID)
}
