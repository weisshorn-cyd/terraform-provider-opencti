resource "opencti_marking_definition" "marking_definitions" {
  for_each = { for m in var.markings : m.definition => m }

  definition_type = each.value.definitionType
  definition      = each.value.definition
  x_opencti_order = each.value.xOpenctiOrder
  x_opencti_color = each.value.xOpenctiColor
}

output "output_marking_definitions" {
  value = {
    for m in opencti_marking_definition.marking_definitions :
    m.definition => m if m.definition != null
  }
}
