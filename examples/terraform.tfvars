case_templates = [
  {
    name        = "Test",
    description = "Case template for testing purposes",
  },
]

groups = [
  {
    name        = "Analyst",
    description = "Group for the analyst - can view and edit knowledge",
    roles = [
      "Analyst",
    ],
    allowedMarking = [
      "TLP:RED",
    ],
    autoNewMarking     = true,
    defaultAssignation = false,
    maxConfidenceLevel = 100,
  },
  {
    name        = "AnalystLabelEditor",
    description = "Group for analyst that can manage labels and definitions",
    roles = [
      "Analyst",
      "LabelEditor",
    ],
    allowedMarking = [
      "TLP:RED",
    ],
    autoNewMarking     = true,
    defaultAssignation = false,
    maxConfidenceLevel = 100,
  },
  {
    name        = "Manager",
    description = "Group for users that can manage labels, definitions and markings",
    roles = [
      "Analyst",
      "LabelEditor",
      "MarkingEditor",
    ],
    allowedMarking = [
      "TLP:RED",
    ],
    autoNewMarking     = true,
    defaultAssignation = false,
    maxConfidenceLevel = 100,
  },
]

markings = [
  {
    definitionType = "TLP",
    definition     = "TLP:CLEAR",
    xOpenctiOrder  = 1,
    xOpenctiColor  = "#ffffff",
  },
  {
    definitionType = "TLP",
    definition     = "TLP:GREEN",
    xOpenctiOrder  = 2,
    xOpenctiColor  = "#2e7d32",
  },
  {
    definitionType = "TLP",
    definition     = "TLP:AMBER",
    xOpenctiOrder  = 3,
    xOpenctiColor  = "#d84315",
  },
  {
    definitionType = "TLP",
    definition     = "TLP:AMBER+STRICT",
    xOpenctiOrder  = 4,
    xOpenctiColor  = "#d84315",
  },
  {
    definitionType = "TLP",
    definition     = "TLP:RED",
    xOpenctiOrder  = 5,
    xOpenctiColor  = "#c62828",
  },
  {
    definitionType = "TLP",
    definition     = "TLP:TEST",
    xOpenctiOrder  = 2,
    xOpenctiColor  = "#104d13",
  },
]

roles = [
  {
    name = "Analyst",
    capabilities = [
      "KNOWLEDGE",
      "KNOWLEDGE_KNUPDATE",
      "KNOWLEDGE_KNPARTICIPATE",
      "KNOWLEDGE_KNUPDATE_KNDELETE",
      "KNOWLEDGE_KNUPLOAD",
      "KNOWLEDGE_KNASKIMPORT",
      "KNOWLEDGE_KNGETEXPORT",
      "KNOWLEDGE_KNGETEXPORT_KNASKEXPORT",
      "KNOWLEDGE_KNENRICHMENT",
      "EXPLORE",
      "EXPLORE_EXUPDATE",
      "EXPLORE_EXUPDATE_EXDELETE",
    ]
  },
  {
    name = "LabelEditor",
    capabilities = [
      "SETTINGS_SETLABELS",
    ]
  },
  {
    name = "MarkingEditor",
    capabilities = [
      "SETTINGS_SETMARKINGS",
    ]
  },
]

status_templates = [
  {
    name  = "NEW",
    color = "#ff9800",
    workflows = [
      {
        entity = "Case-Incident",
        order  = 0,
      },
      {
        entity = "Case-Rfi",
        order  = 0,
      },
      {
        entity = "Task",
        order  = 0,
      },
    ]
  },
  {
    name  = "IN_PROGRESS",
    color = "#5c7bf5",
    workflows = [
      {
        entity = "Case-Incident",
        order  = 1,
      },
      {
        entity = "Case-Rfi",
        order  = 1,
      },
      {
        entity = "Task",
        order  = 1,
      },
    ]
  },
  {
    name  = "ON_HOLD",
    color = "#d0021b",
    workflows = [
      {
        entity = "Case-Incident",
        order  = 2,
      },
      {
        entity = "Case-Rfi",
        order  = 2,
      },
      {
        entity = "Task",
        order  = 2,
      },
    ]
  },
  {
    name  = "NOT_APPLICABLE",
    color = "#9b9b9b",
    workflows = [
      {
        entity = "Case-Incident",
        order  = 3,
      },
      {
        entity = "Case-Rfi",
        order  = 3,
      },
      {
        entity = "Task",
        order  = 3,
      },
    ]
  },
  {
    name  = "CLOSED",
    color = "#417505",
    workflows = [
      {
        entity = "Case-Incident",
        order  = 4,
      },
      {
        entity = "Case-Rfi",
        order  = 4,
      },
      {
        entity = "Task",
        order  = 4,
      },
    ]
  },
]

task_templates_test = [
  {
    name        = "1 - Turn on coffee machine",
    description = "Someone needs to turn the coffee machine on.",
  },
  {
    name        = "2 - Put coffee beans",
    description = "Someone needs to put coffee beans",
  },
  {
    name        = "3 - Make coffee",
    description = "",
  },
  {
    name        = "4 - Enjoy :)",
    description = "or not, I don't care.",
  },
]

users = [
  {
    name       = "[C] abc"
    user_email = "abc@test.io"
    groups = [
      "Analyst",
      "AnalystLabelEditor",
    ]
  },
  {
    name       = "[C] def"
    user_email = "def@test.io"
    groups = [
      "Analyst",
    ]
    user_confidence_level = {
      max_confidence = 90
      overrides = [
        {
          entity_type = "Indicator"
          max_confidence = 80
        },
        {
          entity_type = "Malware"
          max_confidence = 70
        }
      ]
    }
  },
  {
    name       = "[C] ghi"
    user_email = "ghi@test.io"
    groups = [
      "Manager",
    ]
    user_confidence_level = {
      max_confidence = 10
      overrides = []
    }
  },
]

vocabularies = [
  {
    name        = "test",
    description = "Test report",
    category    = "report_types_ov",
  },
  {
    name        = "test",
    description = "Test request for information",
    category    = "request_for_information_types_ov",
  },
]
