resource "bitwarden-sm_secret" "test" {
    key             = "secretKey"
    value           = "secretValue"
    note            = "notes"
    project_id      = "projectID"
    organization_id = "organizationId"
}

output "secret" {
    value = {
        id  = data.bitwarden-sm_secret.secret.id
        key = data.bitwarden-sm_secret.secret.key
        # The actual secret value is marked sensitive
        # value          = data.bitwarden-sm_secret.secret.value
        note            = data.bitwarden-sm_secret.secret.note
        project_id      = data.bitwarden-sm_secret.secret.project_id
        organization_id = data.bitwarden-sm_secret.secret.organization_id
        creation_date   = data.bitwarden-sm_secret.secret.creation_date
        revision_date   = data.bitwarden-sm_secret.secret.revision_date
    }
}
