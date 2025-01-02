package dto

import (
	"github.com/chenparty/gog/example/internal/model"
)

type UserInfo struct {
	model.User
	RoleNo string `json:"role_no"`
}
