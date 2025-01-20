# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    opencti = {
      source  = "registry.terraform.io/weisshorn-cyd/opencti"
      version = "0.1.0"
    }
    #vault = {
    #  source = "registry.terraform.io/hashicorp/vault"
    #  version = "4.3.0"
    #}
  }
}

# opencti settings
provider "opencti" {
  url   = "http://localhost:8080"
  token = var.opencti_token
}
