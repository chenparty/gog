package httpcli

import (
	"context"
	"errors"
	"github.com/chenparty/gog/zlog"
	"github.com/chenparty/gog/zlog/ginplugin"
	"github.com/go-resty/resty/v2"
	"net/url"
	"time"
)

var defaultClient *resty.Client
var defaultTimeout = 10 * time.Second

func init() {
	defaultClient = resty.New()
	defaultClient.SetTimeout(defaultTimeout)
	defaultClient.SetHeader("User-Agent", "httpcli/1.0")
}

// RequestOptions 请求选项函数类型
type RequestOptions func(*requestOptions)

// requestOptions 请求选项配置
type requestOptions struct {
	timeout time.Duration
}

// WithTimeout 设置请求超时时间，覆盖默认的 10s 超时
func WithTimeout(timeout time.Duration) RequestOptions {
	return func(o *requestOptions) {
		o.timeout = timeout
	}
}

// getClient 根据选项获取合适的客户端
// 当指定了自定义超时时，创建临时客户端以避免影响全局共享客户端
func getClient(opts *requestOptions) *resty.Client {
	if opts.timeout > 0 {
		c := resty.New()
		c.SetTimeout(opts.timeout)
		c.SetHeader("User-Agent", "httpcli/1.0")
		return c
	}
	return defaultClient
}

func PostJson(ctx context.Context, reqUrl string, header map[string]string, body any, opts ...RequestOptions) (statusCode int, respBody []byte, err error) {
	o := &requestOptions{}
	for _, opt := range opts {
		opt(o)
	}

	if reqUrl == "" {
		zlog.Error().Ctx(ctx).Msg("PostJson: empty URL provided")
		err = errors.New("empty URL")
		return
	}
	if _, err = url.Parse(reqUrl); err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("url", reqUrl).Msg("PostJson: invalid URL")
		err = errors.New("invalid URL")
		return
	}
	if header == nil {
		header = make(map[string]string)
	}
	header["Content-Type"] = "application/json"
	header[ginplugin.HeaderRequestID] = zlog.TraceIDFromContext(ctx)

	c := getClient(o)
	req := c.R().
		SetContext(ctx).
		SetHeaders(header).
		SetBody(body)

	zlog.Info().Ctx(ctx).Str("url", reqUrl).
		Any("body", body).
		Msg("PostJson-Request")

	resp, err := req.Post(reqUrl)
	if err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("url", reqUrl).Msg("PostJson error")
		return
	}

	respBody = resp.Body()
	statusCode = resp.StatusCode()

	zlog.Info().Ctx(ctx).Str("url", reqUrl).
		Str("status", resp.Status()).
		Dur("time", resp.Time()).
		Str("body", string(respBody)).
		Msg("PostJson-Response")

	return
}

func Get(ctx context.Context, reqUrl string, header map[string]string, queryParam map[string]string, opts ...RequestOptions) (statusCode int, respBody []byte, err error) {
	o := &requestOptions{}
	for _, opt := range opts {
		opt(o)
	}

	if reqUrl == "" {
		zlog.Error().Ctx(ctx).Msg("Get: empty URL provided")
		err = errors.New("empty URL")
		return
	}
	if _, err = url.Parse(reqUrl); err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("url", reqUrl).Msg("Get: invalid URL")
		err = errors.New("invalid URL")
		return
	}
	if header == nil {
		header = make(map[string]string)
	}
	header[ginplugin.HeaderRequestID] = zlog.TraceIDFromContext(ctx)

	c := getClient(o)
	req := c.R().
		SetContext(ctx).
		SetHeaders(header).
		SetQueryParams(queryParam)

	zlog.Info().Ctx(ctx).Str("url", reqUrl).
		Any("body", queryParam).
		Msg("Get-Request")

	resp, err := req.Get(reqUrl)
	if err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("url", reqUrl).Msg("Get error")
		return
	}

	respBody = resp.Body()
	statusCode = resp.StatusCode()

	zlog.Info().Ctx(ctx).Str("url", reqUrl).
		Str("status", resp.Status()).
		Dur("time", resp.Time()).
		Str("body", string(respBody)).
		Msg("Get-Response")

	return
}
