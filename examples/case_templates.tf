resource "opencti_case_template" "case_templates" {
  name        = "Test case template"
  description = "Case template for Test"
  tasks       = [for i in resource.opencti_task_template.task_templates_test : i.id]
}

output "output_case_templates" {
  value = {
    name        = opencti_case_template.case_templates.name
    description = opencti_case_template.case_templates.description
    tasks       = opencti_case_template.case_templates.tasks
  }
}
