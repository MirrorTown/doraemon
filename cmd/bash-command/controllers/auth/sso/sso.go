package sso

import (
	"github.com/Qihoo360/doraemon/cmd/bash-command/controllers/auth"
	"github.com/Qihoo360/doraemon/cmd/bash-command/models"
)

type SSoAuth struct{}

func init() {
	auth.Register(models.AuthTypeSso, &SSoAuth{})
}

func (*SSoAuth) Authenticate(m models.AuthModel) (*models.User, error) {
	username := m.Username
	user, _ := models.UserModel.GetUserByName(username)

	return user, nil
}
