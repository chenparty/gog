package user

import (
	"context"
	"github.com/chenparty/gog/client/mysqlcli"
	"github.com/chenparty/gog/example/internal/app/dto"
	"github.com/chenparty/gog/example/internal/model"
)

func (_ _defaultDao) UserInfo(ctx context.Context, uid string) (userInfo *dto.UserInfo, err error) {
	err = mysqlcli.DB(ctx).Model(&model.User{}).
		Joins("left join role on user.role_no = role.no").
		Where("user.id = ?", uid).
		Select("user.*, role.name").
		Take(&userInfo).Error
	return
}

func (_ _defaultDao) AddUser(ctx context.Context, user *model.User) error {
	return mysqlcli.DB(ctx).Create(user).Error
}
