terraform {
  required_providers {
    bitwarden-sm = {
      source = "registry.terraform.io/bitwarden/bitwarden-sm"
    }
  }
}

provider "bitwarden-sm" {
  api_url         = "https://api.bitwarden.com"
  identity_url    = "https://identity.bitwarden.com"
  access_token    = "< secret machine account access token >"
  organization_id = "< your organization id >"
}

data "bitwarden-sm_projects" "projects" {}

output "projects" {
  value = data.bitwarden-sm_projects.projects
}

data "bitwarden-sm_list_secrets" "secrets" {}

output "secrets" {
  value = data.bitwarden-sm_list_secrets.secrets
}

data "bitwarden-sm_secret" "secret" {
  id = "< secret id >"
}

output "secret" {
  value = {
    id  = resource.bitwarden-sm_secret.secret.id
    key = resource.bitwarden-sm_secret.secret.key
    # The actual secret value is marked sensitive
    # value         = resource.bitwarden-sm_secret.secret.value
    note            = resource.bitwarden-sm_secret.secret.note
    project_id      = resource.bitwarden-sm_secret.secret.project_id
    organization_id = resource.bitwarden-sm_secret.secret.organization_id
    creation_date   = resource.bitwarden-sm_secret.secret.creation_date
    revision_date   = resource.bitwarden-sm_secret.secret.revision_date
  }
}
