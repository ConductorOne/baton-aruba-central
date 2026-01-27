package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	BaseHostField = field.StringField(
		"api-base-host",
		field.WithDescription("The base hostname for the Aruba Central API"),
		field.WithDefaultValue("apigw-uswest5.central.arubanetworks.com"),
	)

	ClientIDField = field.StringField(
		"aruba-central-client-id",
		field.WithDescription("The client ID of the OAuth2 application for the Aruba Central API"),
		field.WithRequired(true),
	)

	ClientSecretField = field.StringField(
		"aruba-central-client-secret",
		field.WithDescription("The client secret of the OAuth2 application for the Aruba Central API"),
		field.WithRequired(true),
	)

	AccessTokenField = field.StringField(
		"access-token",
		field.WithDescription("The access token for the Aruba Central API to be used with refresh token flow"),
	)

	RefreshTokenField = field.StringField(
		"refresh-token",
		field.WithDescription("The refresh token for the Aruba Central API to be used with refresh token flow"),
	)

	UsernameField = field.StringField(
		"username",
		field.WithDescription("The username for the Aruba Central API to be used with code flow"),
	)

	PasswordField = field.StringField(
		"password",
		field.WithDescription("The password for the Aruba Central API to be used with code flow"),
	)

	CustomerIDField = field.StringField(
		"customer-id",
		field.WithDescription("The customer ID for the Aruba Central API to be used with code flow"),
	)

	ConfigurationFields = []field.SchemaField{
		BaseHostField,
		ClientIDField,
		ClientSecretField,
		AccessTokenField,
		RefreshTokenField,
		UsernameField,
		PasswordField,
		CustomerIDField,
	}

	ConfigurationSchema = field.Configuration{
		Fields: ConfigurationFields,
	}
)

//go:generate go run ./gen
var Config = field.NewConfiguration(ConfigurationFields)
