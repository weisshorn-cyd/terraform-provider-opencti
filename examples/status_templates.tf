resource "opencti_status_template" "status_templates" {
  for_each = { for s in var.status_templates : s.name => s }
  name     = each.value.name
  color    = each.value.color
  workflows = [for workflow in each.value.workflows : {
    entity = workflow.entity
    order  = workflow.order
  }]
}

output "output_status_template" {
  value = {
    for s in opencti_status_template.status_templates : s.name => s
  }
}
