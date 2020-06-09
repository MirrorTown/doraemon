package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type UserType int

const (
	DefaultUser UserType = iota
	SystemUser
	APIUser

	TableNameUser = "users"
)

var (
	APIKeyUser = User{
		Id:      0,
		Name:    "OpenAPI",
		Type:    APIUser,
		Display: "OpenAPI",
	}

	AnonymousUser = User{
		Id:      0,
		Name:    "Anonymous",
		Type:    DefaultUser,
		Display: "Anonymous",
	}
)

type userModel struct{}

type User struct {
	Id        int64      `orm:"pk;auto" json:"id,omitempty"`
	Name      string     `orm:"index;unique;size(200)" json:"name,omitempty"`
	Password  string     `orm:"size(255)" json:"-"`
	Salt      string     `orm:"size(32)" json:"-"`
	Email     string     `orm:"size(200)" json:"email,omitempty"`
	Display   string     `orm:"size(200)" json:"display,omitempty"`
	Comment   string     `orm:"type(text)" json:"comment,omitempty"`
	Type      UserType   `orm:"type(integer)" json:"type"`
	Admin     bool       `orm:"default(False)" json:"admin"`
	LastLogin *time.Time `orm:"auto_now_add;type(datetime)" json:"lastLogin,omitempty"`
	LastIp    string     `orm:"size(200)" json:"lastIp,omitempty"`

	Deleted    bool       `orm:"default(false)" json:"deleted,omitempty"`
	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
}

func (*User) TableName() string {
	return TableNameUser
}

func (u *User) GetTypeName() string {
	mapDict := map[UserType]string{
		DefaultUser: "default",
		SystemUser:  "system",
		APIUser:     "api",
	}
	name, ok := mapDict[u.Type]
	if ok == false {
		return ""
	}
	return name
}

func (*userModel) AddUser(m *User) (id int64, err error) {
	id, err = Ormer().Insert(m)
	if err != nil {
		return
	}

	//err = addDefaultNamespace(m)
	//if err != nil {
	//	return
	//}
	return id, nil
}

func (*userModel) EnsureUser(m *User) (*User, error) {
	oldUser := &User{Name: m.Name}
	err := Ormer().Read(oldUser, "Name")
	if err != nil {
		if err == orm.ErrNoRows {
			_, err := UserModel.AddUser(m)
			if err != nil {
				return nil, err
			}
			oldUser = m
		} else {
			return nil, err
		}
	} else {
		oldUser.Email = m.Email
		oldUser.Display = m.Display
		oldUser.LastLogin = m.LastLogin
		oldUser.LastIp = m.LastIp
		_, err := Ormer().Update(oldUser)
		if err != nil {
			return nil, err
		}
	}

	//err = addDefaultNamespace(oldUser)
	return oldUser, err
}

func (*userModel) GetUserByName(name string) (v *User, err error) {
	v = &User{Name: name}
	if err = Ormer().Read(v, "Name"); err != nil {
		return nil, err
	}
	return v, nil
}

func (*userModel) GetUserDetail(name string) (user *User, err error) {
	return nil, nil
}
