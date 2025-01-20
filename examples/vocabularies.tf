# Copyright (c) HashiCorp, Inc.


resource "opencti_vocabulary" "vocabularies" {
  for_each    = { for v in var.vocabularies : format("%s_%s", v.name, v.category) => v }
  name        = each.value.name
  description = each.value.description
  category    = each.value.category
}

output "output_vocabulary" {
  value = {
    for v in opencti_vocabulary.vocabularies : format("%s_%s", v.name, v.category) => v
  }
}
