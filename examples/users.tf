resource "opencti_user" "users" {
  for_each = { for u in var.users : u.name => u }

  name                  = each.value.name
  user_email            = each.value.user_email
  groups                = each.value.groups
  user_confidence_level = each.value.user_confidence_level

  depends_on = [opencti_group.groups]
}

output "opencti_users" {
  value = {
    for u in opencti_user.users : u.name => u
  }
  sensitive = true
}
