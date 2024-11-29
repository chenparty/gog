package gormcli

import (
	"fmt"
	"gog/zlog"
	"gog/zlog/gormplugin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

var db *gorm.DB

type MysqlOption struct {
	Host   string
	Port   string
	User   string
	Pwd    string
	DBName string

	Silent                    bool
	ParameterizedQueries      bool
	IgnoreRecordNotFoundError bool
	SlowThreshold             time.Duration
}

func (o MysqlOption) NewMysqlClient() {
	addr := fmt.Sprintf("tcp(%s:%s)", o.Host, o.Port)
	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local", o.User, o.Pwd, addr, o.DBName)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",   // 表名前缀，`User` 的表名应该是 `t_users`
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
		Logger: gormplugin.New(gormplugin.Config{
			Silent:                    o.Silent,
			SlowThreshold:             o.SlowThreshold,
			ParameterizedQueries:      o.ParameterizedQueries,
			IgnoreRecordNotFoundError: o.IgnoreRecordNotFoundError,
		}),
	})
	if err != nil {
		zlog.Error().Err(err).Msg("DB Open failed")
	}
	zlog.Info().Msg("DB Open success")
}

// DB 使用时请务必使用WithContext函数传递上下文
func DB() *gorm.DB {
	return db
}
