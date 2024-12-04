package miniocli

import (
	"context"
	"github.com/chenparty/gog/zlog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"time"
)

var minioClient *minio.Client

type Options struct {
	AccessKeyID     string
	SecretAccessKey string

	useSSL bool
}

type Option func(*Options)

// Connect 连接minio
func Connect(addr string, options ...Option) {
	opts := Options{}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	minioOptions := minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKeyID, opts.SecretAccessKey, ""),
		Secure: opts.useSSL,
	}
	var err error
	minioClient, err = minio.New(addr, &minioOptions)
	if err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("minio连接失败")
		panic(err)
	}
	zlog.Info().Str("addr", addr).Msg("minio连接成功")
}

// WithAccess 设置访问密钥
func WithAccess(id, secret string) Option {
	return func(options *Options) {
		options.AccessKeyID = id
		options.SecretAccessKey = secret
	}
}

// WithSSL 使用SSL
func WithSSL(useSSL bool) Option {
	return func(options *Options) {
		options.useSSL = useSSL
	}
}

// CheckBucket 检查桶是否存在，不存在则创建
func CheckBucket(ctx context.Context, bucketName string) (err error) {
	// 检查桶是否存在
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return
	}
	if !exists {
		// 创建新桶
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}
	return
}

// PutObject 上传对象
func PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (err error) {
	options := minio.PutObjectOptions{ContentType: contentType}
	_, err = minioClient.PutObject(ctx, bucketName, objectName, reader, objectSize, options)
	return
}

// DelObject 删除对象
func DelObject(ctx context.Context, bucketName, objName string) (err error) {
	err = minioClient.RemoveObject(ctx, bucketName, objName, minio.RemoveObjectOptions{})
	return
}

// PreSignedGetObject 生成带有授权访问的临时URL
func PreSignedGetObject(ctx context.Context, bucketName, objName string, expiration time.Duration) (url string, err error) {
	preSignedURL, err := minioClient.PresignedGetObject(ctx, bucketName, objName, expiration, nil)
	if err != nil {
		return
	}
	url = preSignedURL.String()
	return
}

// GetObject 获取对象
func GetObject(ctx context.Context, bucketName, objName string) (obj *minio.Object, err error) {
	obj, err = minioClient.GetObject(ctx, bucketName, objName, minio.GetObjectOptions{})
	return
}
