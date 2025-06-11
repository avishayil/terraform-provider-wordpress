# Common configuration to be used with examples

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    wordpress = {
      source  = "avishayil/wordpress"
      version = "~> 0.1.0"
    }
  }
}