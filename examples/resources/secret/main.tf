provider "bitwarden-sm" {
  api_url         = var.api_url
  identity_url    = var.identity_url
  organization_id = var.organization_id
  access_token    = var.access_token
}

resource "bitwarden-sm_secret" "secret" {
  key        = var.key
  value      = var.value
  note       = var.note
  project_id = var.project_id
}

resource "local_file" "example" {
  content  = <<EOF
{
    "db_username_secret": {
        "id" : "${resource.bitwarden-sm_secret.secret.id}",
        "key" : "${resource.bitwarden-sm_secret.secret.key}",
        "value" : "${resource.bitwarden-sm_secret.secret.value}",
        "note" : "${resource.bitwarden-sm_secret.secret.note}",
        "project_id" : "${resource.bitwarden-sm_secret.secret.project_id}",
        "organization_id" : "${resource.bitwarden-sm_secret.secret.organization_id}",
        "creation_date" : "${resource.bitwarden-sm_secret.secret.creation_date}",
        "revision_date" : "${resource.bitwarden-sm_secret.secret.revision_date}"
    }
}
EOF
  filename = "${path.module}/output.json"
}

output "secret" {
  value = {
    id              = resource.bitwarden-sm_secret.secret.id
    key             = resource.bitwarden-sm_secret.secret.key
    note            = resource.bitwarden-sm_secret.secret.note
    project_id      = resource.bitwarden-sm_secret.secret.project_id
    organization_id = resource.bitwarden-sm_secret.secret.organization_id
    creation_date   = resource.bitwarden-sm_secret.secret.creation_date
    revision_date   = resource.bitwarden-sm_secret.secret.revision_date
  }
}
