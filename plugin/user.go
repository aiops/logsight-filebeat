package plugin

import (
	"fmt"
	"github.com/google/uuid"
)

type User struct {
	Id       *uuid.UUID
	Email    string
	Password string
}

func (u User) String() string {
	return fmt.Sprintf(`{"id": "%v", "email": "%v"}`, u.Id, u.Email)
}

type UserApi struct {
	loginApi *LoginApi
}

func (u *UserApi) login(email string, password string) (*User, error) {
	loginReq := LoginRequest{Email: email, Password: password}
	if loginResp, err := u.loginApi.Login(loginReq); err != nil {
		return nil, err
	} else {
		return &User{Id: &loginResp.User.Id, Email: email, Password: password}, nil
	}
}
