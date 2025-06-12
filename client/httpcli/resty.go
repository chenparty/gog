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

var client *resty.Client

func init() {
	client = resty.New().
		SetTimeout(10*time.Second).
		SetHeader("User-Agent", "httpcli/1.0")
}

func PostJson(ctx context.Context, reqUrl string, header map[string]string, body any) (statusCode int, respBody []byte, err error) {
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

	req := client.R().
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

func Get(ctx context.Context, reqUrl string, header map[string]string, queryParam map[string]string) (statusCode int, respBody []byte, err error) {
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

	req := client.R().
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
