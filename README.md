![Baton Logo](./docs/images/baton-logo.png)

# `baton-aruba-central` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-aruba-central.svg)](https://pkg.go.dev/github.com/conductorone/baton-aruba-central) ![main ci](https://github.com/conductorone/baton-aruba-central/actions/workflows/main.yaml/badge.svg)

`baton-aruba-central` is a connector for ArubaCentral built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the ArubaCentral API, to sync data about users, roles and groups. 

Check out [Baton](https://github.com/conductorone/baton) to learn more about the project in general.

# Prerequisites

To be able to work with the connector, you need to have an API Gateway set up in Aruba Central instance. This API Gateway gives you an ability to create client applications along with a token. You can set this up in the Aruba Central UI by going to `Organization` in the main left sidebar menu, then choosing `Platform Integration` tab at the top. Here you can see the API Gateway window with link `Rest API` which will take you to another page where you can manage your API client applications. You can also find here your base hostname on which you can find documentation and on which you can acess the Rest API. You can also use the following [link](https://developer.arubanetworks.com/aruba-central/docs/api-gateway) to get more information about the API Gateway. 

Connector enables two ways of authentication with the Aruba Central API. Both ways require you to have a client ID and client secret. Once you have the API Gateway set up, you can create a new client application and get the client ID and client secret. 

You can then use along the client ID and client secret, your username, password and customer ID to authenticate through OAuth Code Flow that automatically retrieves the refresh token and access token. Customer ID is available in the Aruba Central UI in the top right corner of the screen.

Or you can use along the client ID and client secret, the refresh token and access token to authenticate through OAuth Refresh Token Flow. Access token and refresh token are accessible along the client applications in the API Gateway. You have to download the token to retrieve them. Access token is valid for 2 hours and refresh token is valid for 14 days. After you invalidate the access token, connector automatically refreshes it using the refresh token. This refresh operation invalidates the previous refresh token and generates a new one.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-aruba-central

BATON_API_BASE_HOST=api-base-host BATON_ARUBA_CENTRAL_CLIENT_ID=aruba-central-client-id BATON_ARUBA_CENTRAL_CLIENT_SECRET=aruba-central-client-secret BATON_REFRESH_TOKEN=refresh-token BATON_ACCESS_TOKEN=access-token baton-aruba-central
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_API_BASE_HOST=api-base-host BATON_ARUBA_CENTRAL_CLIENT_ID=aruba-central-client-id BATON_ARUBA_CENTRAL_CLIENT_SECRET=aruba-central-client-secret BATON_USERNAME=username BATON_PASSWORD=password BATON_CUSTOMER_ID=customer-id ghcr.io/conductorone/baton-aruba-central:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-aruba-central/cmd/baton-aruba-central@main

BATON_API_BASE_HOST=api-base-host BATON_ARUBA_CENTRAL_CLIENT_ID=aruba-central-client-id BATON_ARUBA_CENTRAL_CLIENT_SECRET=aruba-central-client-secret BATON_REFRESH_TOKEN=refresh-token BATON_ACCESS_TOKEN=access-token baton-aruba-central
baton resources
```

# Data Model

`baton-aruba-central` will fetch information about the following ArubaCentral resources:

- Users
- Roles
- Groups

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-aruba-central` Command Line Usage

```
baton-aruba-central

Usage:
  baton-aruba-central [flags]
  baton-aruba-central [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --access-token string                  The access token for the Aruba Central API to be used with refresh token flow. ($BATON_ACCESS_TOKEN)
      --api-base-host string                 The base hostname for the Aruba Central API. ($BATON_API_BASE_HOST) (default "apigw-uswest5.central.arubanetworks.com")
      --aruba-central-client-id string       The client ID of the OAuth2 application for the Aruba Central API. ($BATON_ARUBA_CENTRAL_CLIENT_ID)
      --aruba-central-client-secret string   The client secret of the OAuth2 application for the Aruba Central API. ($BATON_ARUBA_CENTRAL_CLIENT_SECRET)
      --client-id string                     The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string                 The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --customer-id string                   The customer ID for the Aruba Central API to be used with code flow. ($BATON_CUSTOMER_ID)
  -f, --file string                          The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                                 help for baton-aruba-central
      --log-format string                    The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string                     The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --password string                      The password for the Aruba Central API to be used with code flow. ($BATON_PASSWORD)
  -p, --provisioning                         This must be set in order for provisioning actions to be enabled. ($BATON_PROVISIONING)
      --refresh-token string                 The refresh token for the Aruba Central API to be used with refresh token flow. ($BATON_REFRESH_TOKEN)
      --username string                      The username for the Aruba Central API to be used with code flow. ($BATON_USERNAME)
  -v, --version                              version for baton-aruba-central

Use "baton-aruba-central [command] --help" for more information about a command.
```
