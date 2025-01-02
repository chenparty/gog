package user

type userInfoParam struct {
	ID string `form:"id" binding:"required"`
}

type addUserParam struct {
	Name string `json:"name" binding:"required"`
}
