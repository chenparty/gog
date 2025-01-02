package user

import (
	"context"
	"github.com/chenparty/gog/example/internal/app/mq/resp"
	"github.com/chenparty/gog/example/internal/model"
)

func (s *_defaultService) UserInfo(ctx context.Context, uid string) *resp.Output {
	userInfo, err := s.userDao.UserInfo(ctx, uid)
	if err != nil {
		return resp.DBErr.Output().WithMsg("用户查询失败")
	}
	return resp.OK.Output().WithData(userInfo)
}

func (s *_defaultService) AddUser(ctx context.Context, username string) *resp.Output {
	user := model.User{
		Name: username,
	}
	err := s.userDao.AddUser(ctx, &user)
	if err != nil {
		return resp.DBErr.Output().WithMsg("用户创建失败")
	}
	return resp.OK.Output()
}
