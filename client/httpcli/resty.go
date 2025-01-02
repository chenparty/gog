package httpcli

import (
	"context"
	"github.com/chenparty/gog/zlog"
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

func PostJson(ctx context.Context, url string, body any) {
	resp, err := client.R().SetHeader("Content-Type", "application/json").SetBody(body).Post(url)
	if err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("url", url).Msg("post error")
		return
	}
	zlog.Info().Ctx(ctx).Str("url", url).Str("status", resp.Status()).Dur("time", resp.Time()).Str("body", string(resp.Body())).Msg("请求结果")
}

func Get(ctx context.Context, url string, param map[string]string) {
	resp, err := client.R().SetQueryParams(param).Get(url)
	if err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("url", url).Msg("get error")
		return
	}
	zlog.Info().Ctx(ctx).Str("url", url).Str("status", resp.Status()).Dur("time", resp.Time()).Any("param", param).Msg("请求结果")
}
