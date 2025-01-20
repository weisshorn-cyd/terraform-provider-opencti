# Copyright (c) HashiCorp, Inc.


resource "opencti_group" "groups" {
  for_each = { for g in var.groups : g.name => g }

  name                 = each.value.name
  description          = each.value.description
  roles                = each.value.roles
  allowed_marking      = each.value.allowedMarking
  auto_new_marking     = each.value.autoNewMarking
  default_assignation  = each.value.defaultAssignation
  max_confidence_level = each.value.maxConfidenceLevel

  depends_on = [opencti_role.roles, opencti_marking_definition.marking_definitions]
}

output "output_groups" {
  value = {
    for g in opencti_group.groups : g.name => g
  }
}
