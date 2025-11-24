terraform {
  required_providers {
    opencti = {
      source = "terraform.local/weisshorn-cyd/opencti"
    }
  }
}

# opencti settings
provider "opencti" {
  url   = "http://localhost:8080"
  token = var.opencti_token
}
