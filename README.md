# Terraform Provider Nautobot

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.13.5 or [OpenTofu](https://opentofu.org/docs/intro/install/) >= 1.10.8
- [Go](https://golang.org/doc/install) >= 1.24.X

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the `make` command:

```sh
$ make install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up-to-date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

The provide requires two arguments, `url` and `token`. For the data sources and resources supported, take a look at the [internal/provider](internal/provider) folder. In the next example, we capture the data of all manufacturers and create a new manufacturer "Vendor I". For all arguments that the provider accepts, see its [documentation](docs/index.md)

```hcl
terraform {
  required_providers {
    nautobot = {
      version = "3.0.4"
      source  = "registry.terraform.io/TobiPeterG/nautobot"
    }
  }
}

provider "nautobot" {
  url = "https://demo.nautobot.com/api/"
  token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}

data "nautobot_manufacturers" "all" {}

resource "nautobot_manufacturer" "new" {
  description = "Created with Terraform"
  name    = "Vendor I"
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

There are a few make targets you can leverage:

- `make install`: To compile the provider.
- `make docs`: To generate or update documentation.
- `make local`: Test local version of the provider.
- `make testacc`: To run the full suite of Acceptance tests.

_Note:_ Acceptance tests create real resources, and cost money to run.

```sh
$ make testacc
```

## Credits

This [project](https://github.com/nleiva/terraform-provider-nautobot) started as an exercise for educational purposes by @nleiva during the development of his book "Network Automation with Go". Thank you Nicolas for your effort and collaboration!
