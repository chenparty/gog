package zlog

import "github.com/rs/zerolog"

const CtxTraceIDKey = "trace_id"

// TraceHook 是一个自定义 Hook，用于为每个日志条目添加 trace_id
type TraceHook struct{}

// Run 实现 zerolog.Hook 接口，允许修改日志条目
func (h *TraceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	// 将 trace_id 添加到日志条目
	traceID, _ := e.GetCtx().Value(CtxTraceIDKey).(string)
	e.Str("trace_id", traceID)
}
