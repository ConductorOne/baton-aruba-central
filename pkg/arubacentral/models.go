package arubacentral

import (
	"slices"
)

type User struct {
	Username string `json:"username"`
	Name     struct {
		First string `json:"firstname"`
		Last  string `json:"lastname"`
	} `json:"name"`
	Applications []struct {
		Name string `json:"name"`
		Info []struct {
			Role  string `json:"role"`
			Scope struct {
				Groups []string `json:"groups"`
			} `json:"scope"`
		} `json:"info"`
	} `json:"applications"`
}

func (u *User) ContainsGroup(group string) bool {
	for _, app := range u.Applications {
		for _, info := range app.Info {
			if slices.Contains(info.Scope.Groups, group) {
				return true
			}
		}
	}

	return false
}

type Module struct {
	Name       string `json:"module_name"`
	Permission string `json:"permission"`
}

type Application struct {
	Name       string   `json:"appname"`
	Permission string   `json:"permission"`
	Modules    []Module `json:"modules"`
}

type Role struct {
	RoleName     string        `json:"rolename"`
	Users        []string      `json:"users"`
	NoOfUsers    int           `json:"no_of_users"`
	Permission   string        `json:"permission"`
	Applications []Application `json:"applications"`
}
