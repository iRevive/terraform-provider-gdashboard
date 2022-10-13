# Terraform Provider GDashboard

The provider allows building Grafana panels using Terraform syntax.

## Using the provider

Please, see [provider documentation](https://registry.terraform.io/providers/iRevive/gdashboard/latest/docs).  
The module providers only **data sources**. Each data source emits a JSON that is compatible with Grafana API.    
In order to create dashboard use [Grafana provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs).

## Examples

```terraform
terraform {
  required_providers {
    grafana = {
      source  = "grafana/grafana"
      version = "1.29.0"
    }

    gdashboard = {
      source  = "iRevive/gdashboard"
      version = ">= 0.0.3" # use actual version
    }
  }

  required_version = ">= 1.2.0"
}

data "gdashboard_stat" "status" {
  title       = "Status"
  description = "Shows the status of the container"

  field {
    mappings {
      value {
        value        = "1"
        display_text = "UP"
        color        = "green"
      }

      special {
        match        = "null+nan"
        display_text = "DOWN"
        color        = "red"
      }
    }
  }

  targets {
    prometheus {
      uid     = "prometheus"
      expr    = "up{container_name='container'}"
      instant = true
    }
  }
}

data "gdashboard_dashboard" "dashboard" {
  title = "My dashboard"

  layout {
    row {
      panel {
        size = {
          height = 8
          width  = 10
        }
        source = data.gdashboard_stat.status.json
      }
    }
  }
}

# Define provider
provider "grafana" {
  url  = "https://my.grafana.com" # use your API endpoint
  auth = var.grafana_auth
}

# Create your dashboard
resource "grafana_dashboard" "my_dashboard" {
  config_json = data.gdashboard_dashboard.dashboard.json
}
```

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.18

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
