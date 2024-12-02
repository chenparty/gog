package zlog

import (
	"context"
	"github.com/oklog/ulid/v2"
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

// TraceIDFromContext 从上下文中获取 trace_id
func TraceIDFromContext(ctx context.Context) string {
	traceID, _ := ctx.Value(ctxTraceIDKey).(string)
	return traceID
}

// ContextWithValue 将 trace_id 添加到上下文
func ContextWithValue(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxTraceIDKey, traceID)
}

// NewTraceContext 创建一个新的带有 trace_id 的上下文
func NewTraceContext(ctx context.Context) context.Context {
	return ContextWithValue(ctx, NewTraceID())
}

// NewTraceID 生成一个新的 trace_id
func NewTraceID() string {
	return ulid.Make().String()
}
