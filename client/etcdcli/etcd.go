package etcdcli

import (
	"context"
	"github.com/chenparty/gog/zlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"strings"
	"time"
)

var cli *clientv3.Client

type Locker struct {
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

type Options struct {
	Username string
	Password string

	PingKeyPrefix string
}

type Option func(*Options)

// Connect 连接etcd
func Connect(servers []string, options ...Option) {
	opts := Options{}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	var err error
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   servers,
		DialTimeout: 3 * time.Second,
		Username:    opts.Username,
		Password:    opts.Password,
	})
	if err != nil {
		zlog.Error().Err(err).Str("servers", strings.Join(servers, ",")).Msg("etcd连接失败")
		panic(err)
	}
	// 尝试发送一个请求，检查连接是否成功
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = cli.Get(ctx, opts.PingKeyPrefix+"/ping") // 这里可以尝试获取一个存在的键
	if err != nil {
		zlog.Error().Err(err).Msg("etcd get失败")
		panic(err)
	}
	zlog.Info().Str("servers", strings.Join(servers, ",")).Msg("etcd连接成功")
}

// WithUserAndPass 设置用户名密码
func WithUserAndPass(user, pwd string) Option {
	return func(options *Options) {
		options.Username = user
		options.Password = pwd
	}
}

// WithPingKeyPrefix 设置ping key的前缀
func WithPingKeyPrefix(prefix string) Option {
	return func(options *Options) {
		options.PingKeyPrefix = prefix
	}
}

// Close 关闭连接
func Close() {
	_ = cli.Close()
}

// Put 设置key, ttl为租约期, 单位为秒
func Put(ctx context.Context, key, value string, ttl int64) (err error) {
	if ttl <= 0 {
		_, err = cli.Put(ctx, key, value)
		return
	}
	// 创建一个租约对象
	var lease clientv3.Lease
	lease = clientv3.NewLease(cli)
	// 根据时间，生成一个租约
	var leaseResp *clientv3.LeaseGrantResponse
	leaseResp, err = lease.Grant(ctx, ttl)
	if err != nil {
		return
	}
	// 设置key，并绑定租约
	_, err = cli.Put(ctx, key, value, clientv3.WithLease(leaseResp.ID))
	return
}

// Get 根据key获取value
func Get(ctx context.Context, key string) (val string, isNotExist bool, err error) {
	resp, err := cli.Get(ctx, key)
	if err != nil {
		return
	}
	if len(resp.Kvs) == 0 {
		isNotExist = true
		return
	}
	val = string(resp.Kvs[0].Value)
	return
}

// NewLocker 创建一个锁
func NewLocker(ttl int) (l *Locker, err error) {
	session, err := concurrency.NewSession(cli, concurrency.WithTTL(ttl))
	if err != nil {
		return
	}
	l = &Locker{
		session: session,
	}
	return
}

// Close 调用NewLocker后必须调用此方法
func (l *Locker) Close() (err error) {
	err = l.session.Close()
	return
}

// Lock 阻塞式获取锁，锁被占用时则阻塞等待
func (l *Locker) Lock(ctx context.Context, key string) (err error) {
	l.mutex = concurrency.NewMutex(l.session, key)
	err = l.mutex.Lock(ctx)
	return
}

// TryLock 尝试式获取锁，锁被占用时则立即返回err
func (l *Locker) TryLock(ctx context.Context, key string) (err error) {
	l.mutex = concurrency.NewMutex(l.session, key)
	err = l.mutex.TryLock(ctx)
	return
}

// UnLock 解锁
func (l *Locker) UnLock(ctx context.Context) (err error) {
	err = l.mutex.Unlock(ctx)
	return
}

// NewLockerAndLock 新建锁+锁key
func NewLockerAndLock(ctx context.Context, lockKey string, ttl int) (lock *Locker, err error) {
	lock, err = NewLocker(ttl)
	if err != nil {
		return
	}
	err = lock.Lock(ctx, lockKey)
	return
}

// UnlockAndClose 解锁和关锁
func UnlockAndClose(ctx context.Context, lock *Locker) (err error) {
	err = lock.UnLock(ctx)
	if err != nil {
		return
	}
	err = lock.Close()
	return
}
