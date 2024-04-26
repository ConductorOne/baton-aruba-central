package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	BaseHost          string `mapstructure:"api-base-host"`
	ArubaClientID     string `mapstructure:"aruba-central-client-id"`
	ArubaClientSecret string `mapstructure:"aruba-central-client-secret"`
	AccessToken       string `mapstructure:"access-token"`
	RefreshToken      string `mapstructure:"refresh-token"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	CustomerID        string `mapstructure:"customer-id"`
}

func (cfg *config) ShouldUseOAuth2CodeFlow() bool {
	return cfg.Username != "" && cfg.Password != "" && cfg.CustomerID != ""
}

func (cfg *config) ShouldUseOAuth2RefreshTokenFlow() bool {
	return cfg.AccessToken != "" && cfg.RefreshToken != ""
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.ArubaClientID == "" || cfg.ArubaClientSecret == "" {
		return status.Errorf(codes.InvalidArgument, "aruba-central-client-id and aruba-central-client-secret are required, use --help for more information")
	}

	if !cfg.ShouldUseOAuth2CodeFlow() && !cfg.ShouldUseOAuth2RefreshTokenFlow() {
		return status.Errorf(codes.InvalidArgument, "either username, password, and customer-id or access-token and refresh-token are required, use --help for more information")
	}

	return nil
}

func cmdFlags(cmd *cobra.Command) {
	// api base host - default region is US West 5 (more information about other regions:
	// https://developer.arubanetworks.com/aruba-central/docs/api-oauth-access-token#table-domain-urls-for-api-gateway-access
	cmd.PersistentFlags().String("api-base-host", "apigw-uswest5.central.arubanetworks.com", "The base hostname for the Aruba Central API. ($BATON_API_BASE_HOST)")
	cmd.PersistentFlags().String("aruba-central-client-id", "", "The client ID of the OAuth2 application for the Aruba Central API. ($BATON_ARUBA_CENTRAL_CLIENT_ID)")
	cmd.PersistentFlags().String("aruba-central-client-secret", "", "The client secret of the OAuth2 application for the Aruba Central API. ($BATON_ARUBA_CENTRAL_CLIENT_SECRET)")

	// OAuth2 Refresh token flow
	cmd.PersistentFlags().String("access-token", "", "The access token for the Aruba Central API to be used with refresh token flow. ($BATON_ACCESS_TOKEN)")
	cmd.PersistentFlags().String("refresh-token", "", "The refresh token for the Aruba Central API to be used with refresh token flow. ($BATON_REFRESH_TOKEN)")

	// OAuth2 Code flow
	cmd.PersistentFlags().String("username", "", "The username for the Aruba Central API to be used with code flow. ($BATON_USERNAME)")
	cmd.PersistentFlags().String("password", "", "The password for the Aruba Central API to be used with code flow. ($BATON_PASSWORD)")
	cmd.PersistentFlags().String("customer-id", "", "The customer ID for the Aruba Central API to be used with code flow. ($BATON_CUSTOMER_ID)")
}
