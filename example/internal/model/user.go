package model

// User 用户表
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Sex  string `json:"sex"`
}

func (u *User) TableName() string {
	return "user"
}
