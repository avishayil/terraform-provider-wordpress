# Terraform Provider for WordPress

This is a [Terraform](https://www.terraform.io) provider for managing WordPress plugins using [WP-CLI](https://wp-cli.org/) over SSH or Docker.  
It is built using the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

## Features

- Manage WordPress plugins (install, activate, deactivate, delete) via Terraform
- Connect to remote WordPress instances using SSH or Docker
- Supports custom WordPress paths and root access for WP-CLI

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23
- A running WordPress instance accessible via SSH or Docker (see [docker-compose.yml](./docker-compose.yml) for a local setup)

## Building the Provider

1. Clone the repository:
    ```sh
    git clone https://github.com/avishayil/terraform-provider-wordpress.git
    cd terraform-provider-wordpress
    ```
2. Build the provider:
    ```sh
    go install
    ```

## Using the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    wordpress = {
      source  = "avishayil/wordpress"
      version = "0.1.0"
    }
  }
}

provider "wordpress" {
  ssh_target  = "docker:wordpress"
  remote_path = "/var/www/html"
  allow_root  = true
}

resource "wordpress_plugin" "example" {
  name   = "hello-dolly"
  active = true
}
```

- `ssh_target`: The SSH target for remote WordPress execution. E.g., `docker:container-name` or `user@host`.
- `remote_path`: The path to the WordPress installation on the remote system.
- `allow_root`: (Optional) Whether to add `--allow-root` to WP-CLI commands.

## Developing the Provider

1. Install [Go](https://golang.org/doc/install) (see [Requirements](#requirements)).
2. Build the provider:
    ```sh
    go install
    ```
3. Run acceptance tests (requires Docker and a local WordPress environment):
    ```sh
    make testacc
    ```

### Code Generation

To generate or update documentation and code, run:
```sh
make generate
```

## Testing

Acceptance tests require Docker and will spin up a WordPress and MariaDB environment using [docker-compose.yml](./docker-compose.yml):

```sh
docker compose up -d
make testacc
```

*Note:* Acceptance tests create and destroy real resources.

## Contributing

Contributions are welcome! Please open issues or pull requests on [GitHub](https://github.com/avishayil/terraform-provider-wordpress).

## License

This project is licensed under the [MPL-2.0 License](./LICENSE).

---

_This provider is not affiliated with or endorsed by WordPress or Automattic._