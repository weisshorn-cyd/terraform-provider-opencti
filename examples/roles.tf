# Copyright (c) HashiCorp, Inc.


resource "opencti_role" "roles" {
  for_each = { for r in var.roles : r.name => r }

  name         = each.value.name
  capabilities = sort(each.value.capabilities)
}

output "output_roles" {
  value = {
    for r in opencti_role.roles : r.name => r
  }
}
