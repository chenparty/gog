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

// NewTraceID 生成一个新的 trace_id
func NewTraceID() string {
	return ulid.Make().String()
}

// ContextWithValue 基于父context，创建一个带 trace_id 的上下文，受父context生命周期影响，非必要不使用
func ContextWithValue(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxTraceIDKey, traceID)
}

// NewTraceContextWithID 创建一个新的trace context, 使用自定义traceID
func NewTraceContextWithID(traceID string) context.Context {
	if len(traceID) > 0 {
		return ContextWithValue(context.Background(), traceID)
	}
	return ContextWithValue(context.Background(), NewTraceID())
}

// NewTraceContext 创建一个新的trace context
func NewTraceContext() context.Context {
	return ContextWithValue(context.Background(), NewTraceID())
}

// TraceIDFromContext 从上下文中获取 trace_id
func TraceIDFromContext(ctx context.Context) string {
	traceID, _ := ctx.Value(ctxTraceIDKey).(string)
	return traceID
}
