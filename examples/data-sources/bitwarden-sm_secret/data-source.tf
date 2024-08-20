data "bitwarden-sm_secret" "secret" {
  id = "e6a8066c-81e6-428e-bf5d-b1b900fe1b42"
}

output "secret" {
  value = {
    id  = data.bitwarden-sm_secret.secret.id
    key = data.bitwarden-sm_secret.secret.key
    # The actual secret value is marked sensitive and will not be printed to stdout
    # value          = data.bitwarden-sm_secret.secret.value
    note            = resource.bitwarden-sm_secret.secret.note
    project_id      = resource.bitwarden-sm_secret.secret.project_id
    organization_id = resource.bitwarden-sm_secret.secret.organization_id
    creation_date   = resource.bitwarden-sm_secret.secret.creation_date
    revision_date   = resource.bitwarden-sm_secret.secret.revision_date
  }
}
