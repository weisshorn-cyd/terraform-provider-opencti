terraform {
  required_providers {
    opencti = {
      source  = "terraform.local/weisshorn-cyd/opencti"
      version = ">= 0.1.0"
    }
  }
}

# opencti settings
provider "opencti" {
  url   = "http://localhost:8080"
  token = var.opencti_token
}
