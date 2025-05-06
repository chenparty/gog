package httpcli

import (
	"context"
	"github.com/chenparty/gog/zlog"
	"github.com/chenparty/gog/zlog/ginplugin"
	"github.com/go-resty/resty/v2"
	"time"
)

var client *resty.Client

func init() {
	client = resty.New().
		SetTimeout(10 * time.Second).
		SetRetryCount(1).
		SetRetryWaitTime(time.Second)
}

func PostJson(ctx context.Context, url string, header map[string]string, body any) (statusCode int, respBody []byte, err error) {
	if header == nil {
		header = make(map[string]string)
	}
	header["Content-Type"] = "application/json"
	header[ginplugin.HeaderRequestID] = zlog.TraceIDFromContext(ctx)
	req := client.R().
		SetContext(ctx).
		SetHeaders(header).
		SetBody(body)
	zlog.Info().Ctx(ctx).Str("url", url).
		Any("body", body).
		Msg("PostJson-Request")

	resp, err := req.Post(url)
	if err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("url", url).Msg("PostJson error")
		return
	}
	zlog.Info().Ctx(ctx).Str("url", url).
		Str("status", resp.Status()).
		Dur("time", resp.Time()).
		Str("body", string(resp.Body())).
		Msg("PostJson-Response")
	statusCode = resp.StatusCode()
	respBody = resp.Body()
	return
}

func Get(ctx context.Context, url string, header map[string]string, queryParam map[string]string) (statusCode int, respBody []byte, err error) {
	if header == nil {
		header = make(map[string]string)
	}
	header[ginplugin.HeaderRequestID] = zlog.TraceIDFromContext(ctx)
	req := client.R().SetHeaders(header).SetQueryParams(queryParam)
	zlog.Info().Ctx(ctx).Str("url", url).
		Any("body", queryParam).
		Msg("Get-Request")

	resp, err := req.Get(url)
	if err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("url", url).Msg("Get error")
		return
	}
	statusCode = resp.StatusCode()
	respBody = resp.Body()
	zlog.Info().Ctx(ctx).Str("url", url).
		Str("status", resp.Status()).
		Dur("time", resp.Time()).
		Str("body", string(resp.Body())).
		Msg("Get-Response")
	return
}
