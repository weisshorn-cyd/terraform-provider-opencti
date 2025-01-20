# Copyright (c) HashiCorp, Inc.

FROM hashicorp/terraform:1.9

RUN apk add go

# Clone the opencti provider from gitlab
COPY . /usr/local/bin/terraform-provider-opencti
WORKDIR /usr/local/bin/terraform-provider-opencti

# Build the custom provider
RUN CGO_ENABLED=0 go build -o terraform-provider-opencti

RUN mkdir -p /root/.terraform.d/providers/registry.terraform.io/weisshorn-cyd/opencti/0.1.0/linux_amd64/
RUN mv terraform-provider-opencti /root/.terraform.d/providers/registry.terraform.io/weisshorn-cyd/opencti/0.1.0/linux_amd64/

# Terraform config with dev_overrides for opencti provider
COPY terraformrc /root/.terraformrc

ENTRYPOINT ["/bin/terraform"]
