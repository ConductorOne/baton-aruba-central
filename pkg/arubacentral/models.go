package arubacentral

import (
	"google.golang.org/protobuf/types/known/structpb"
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

type ProtoValue interface {
	Marshall() (*structpb.Value, error)
	Unmarshall(*structpb.Value) error
}

type Module struct {
	Name       string `json:"module_name"`
	Permission string `json:"permission"`
}

func (m Module) Marshall() (*structpb.Value, error) {
	fields := make(map[string]*structpb.Value)

	name, err := toStructpbValue(m.Name)
	if err != nil {
		return nil, err
	}
	fields["module_name"] = name

	permission, err := toStructpbValue(m.Permission)
	if err != nil {
		return nil, err
	}
	fields["permission"] = permission

	return structpb.NewStructValue(&structpb.Struct{Fields: fields}), nil
}

func (m *Module) Unmarshall(v *structpb.Value) error {
	if v == nil {
		return nil
	}

	if val := v.Kind.(*structpb.Value_StructValue); val != nil {
		for k, field := range val.StructValue.Fields {
			switch k {
			case "module_name":
				m.Name = field.GetStringValue()
			case "permission":
				m.Permission = field.GetStringValue()
			}
		}
	}

	return nil
}

type Application struct {
	Name       string   `json:"appname"`
	Permission string   `json:"permission"`
	Modules    []Module `json:"modules"`
}

func (a Application) Marshall() (*structpb.Value, error) {
	fields := make(map[string]*structpb.Value)

	name, err := toStructpbValue(a.Name)
	if err != nil {
		return nil, err
	}
	fields["appname"] = name

	permission, err := toStructpbValue(a.Permission)
	if err != nil {
		return nil, err
	}
	fields["permission"] = permission

	var modules []*structpb.Value
	for _, module := range a.Modules {
		m, err := module.Marshall()
		if err != nil {
			return nil, err
		}

		modules = append(modules, m)
	}
	fields["modules"] = structpb.NewListValue(&structpb.ListValue{Values: modules})

	return structpb.NewStructValue(&structpb.Struct{Fields: fields}), nil
}

func (a *Application) Unmarshall(v *structpb.Value) error {
	if v == nil {
		return nil
	}

	if val := v.Kind.(*structpb.Value_StructValue); val != nil {
		for k, field := range val.StructValue.Fields {
			switch k {
			case "appname":
				a.Name = field.GetStringValue()
			case "permission":
				a.Permission = field.GetStringValue()
			case "modules":
				listVal := field.GetListValue()
				for _, item := range listVal.Values {
					var module Module

					err := module.Unmarshall(item)
					if err != nil {
						return err
					}

					a.Modules = append(a.Modules, module)
				}
			}
		}
	}

	return nil
}

type UserString string

func (u UserString) Marshall() (*structpb.Value, error) {
	return structpb.NewStringValue(string(u)), nil
}

func (u *UserString) Unmarshall(v *structpb.Value) error {
	if v == nil {
		return nil
	}

	if val := v.Kind.(*structpb.Value_StringValue); val != nil {
		*u = UserString(val.StringValue)
	}

	return nil
}

type Role struct {
	RoleName     string        `json:"rolename"`
	Users        []UserString  `json:"users"`
	NoOfUsers    int           `json:"no_of_users"`
	Permission   string        `json:"permission"`
	Applications []Application `json:"applications"`
}
