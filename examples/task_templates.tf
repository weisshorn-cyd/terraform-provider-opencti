resource "opencti_task_template" "task_templates_test" {
  for_each    = { for i in var.task_templates_test : i.name => i.description }
  name        = each.key
  description = each.value
}

output "output_task_template_test" {
  value = {
    for i in opencti_task_template.task_templates_test : i.id => {
      name        = i.name
      description = i.description
    }
  }
}
