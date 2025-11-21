variable "opencti_token" {
  description = "OpenCTI admin token for provider"
  type        = string
  sensitive   = true
}

variable "case_templates" {
  description = "A list of opencti case templates to create"
  type = list(object({
    name        = string
    description = string
  }))
}

variable "groups" {
  description = "A list of opencti groups to create"
  type = list(object({
    name               = string
    description        = string
    roles              = list(string)
    allowedMarking     = list(string)
    autoNewMarking     = bool
    defaultAssignation = bool
    maxConfidenceLevel = number
  }))
}

variable "markings" {
  description = "A list of marking definitions to create"
  type = list(object({
    definitionType = string
    definition     = string
    xOpenctiOrder  = number
    xOpenctiColor  = string
  }))
}

variable "roles" {
  description = "A list of opencti roles to create"
  type = list(object({
    name         = string
    capabilities = list(string)
  }))
}

variable "status_templates" {
  description = "A list of opencti status templates to create"
  type = list(object({
    name  = string
    color = string
    workflows = optional(list(object({
      entity = string
      order  = number
    })))
  }))
}

variable "task_templates_test" {
  description = "A list of opencti task templates to create for testing purposes"
  type = list(object({
    name        = string
    description = string
  }))
}

variable "users" {
  description = "A list of opencti users to create"
  type = list(object({
    name       = string
    user_email = string
    groups     = list(string)

    user_confidence_level = optional(object({
      max_confidence = number
      overrides = list(object({
        entity_type = string
        confidence  = number
      }))
    }))
  }))
}

variable "vocabularies" {
  description = "A list of opencti vocabularies to create"
  type = list(object({
    name        = string
    description = string
    category    = string
  }))
}
