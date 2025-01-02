package user

import (
	"context"
	"github.com/chenparty/gog/example/internal/app/dto"
	"github.com/chenparty/gog/example/internal/model"
)

type Dao interface {
	UserInfo(ctx context.Context, uid string) (user *dto.UserInfo, err error)
	AddUser(ctx context.Context, user *model.User) error
}

func NewDao() Dao {
	return &_defaultDao{}
}

type _defaultDao struct{}
