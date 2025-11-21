# Terraform Provider OpenCTI

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25
- [gocti](https://github.com/weisshorn-cyd/gocti)

## Installation

```hcl
terraform {
  required_providers {
    opencti = {
      source = "weisshorn-cyd/opencti"
      version = ">= 0.2.0"
    }
  }
}
```

## Running the provider

1. `$ make prepare-examples`
1. `$ cd examples`
1. `$ TF_VAR_opencti_token="$OPENCTI_ADMIN_TOKEN" terraform init --reconfigure`
1. `$ TF_VAR_opencti_token="$OPENCTI_ADMIN_TOKEN" terraform plan`
1. `$ TF_VAR_opencti_token="$OPENCTI_ADMIN_TOKEN" terraform apply`

## Using the provider

See the [examples](./examples/) folder.

## Developing the Provider

To run the provider in development mode, modify your `.terraformrc` as described [here](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
