# Terraform Provider -  Bitwarden Secrets Manager

_This Terraform provider is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)._

The purpose of this Terraform Provider is to streamline the process of using Bitwarden Secrets Manager within Terraform and OpenTofu, making it more secure and efficient.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.4
- [Go](https://golang.org/doc/install) >= 1.22.5

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install .
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up-to-date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

* work in progress

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

## Acceptance Tests

In order to run the full suite of Acceptance tests, you need to provide the following 2 `.env` files:

1. `.env.local.test`
2. `.env.local.no.access`

Both files should contain the following configuration values.

```text
BW_API_URL="https://your-api-test-endpoint.example.com"
BW_IDENTITY_API_URL="https://your-identity-test-endpoint.example.com"
BW_ACCESS_TOKEN="<your machine account access token >"
BW_ORGANIZATION_ID="< organization id  >"
```

*Important:* The second file `.env.local.no.access` needs to be configured with an access token belonging to a machine account with no project access.

The file [`provider_test.go`](./internal/provider/provider_test.go) uses this file to create the necessary provider configuration.

*Note:* Acceptance tests create real resources, and often cost money to run.

If everything is provided, one can execute all acceptance tests with `make`:

```shell
make testacc
```
