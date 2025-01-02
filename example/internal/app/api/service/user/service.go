package user

import (
	"context"
	"github.com/chenparty/gog/example/internal/app/api/resp"
	"github.com/chenparty/gog/example/internal/app/dao/user"
)

type Service interface {
	UserInfo(ctx context.Context, uid string) *resp.Output
	AddUser(ctx context.Context, username string) *resp.Output
}

func NewService() Service {
	return &_defaultService{userDao: user.NewDao()}
}

type _defaultService struct {
	userDao user.Dao
}
