resource "bitwarden-sm_secret" "db_admin_secret" {
  key = "db_admin_password" #
  # It is not recommended to provide the actual secret value via configuration file
  # By using a terraform variable, users can inject the secret value during runtime via environment variables
  value      = var.value
  note       = "The secret value was provided via terraform configuration"
  project_id = var.project_id
}

# If no secret value is provided, the provider will generate one
# Secret generation is the suggested approach.
resource "bitwarden-sm_secret" "service_account_secret" {
  key        = "db_service_account"
  project_id = var.project_id
}

resource "local_file" "db_admin_credentials" {
  content  = <<EOF
{
    "db_username_secret": {
        "id" : "${resource.bitwarden-sm_secret.db_admin_secret.id}",
        "key" : "${resource.bitwarden-sm_secret.db_admin_secret.key}",
        "value" : "${resource.bitwarden-sm_secret.db_admin_secret.value}",
        "note" : "${resource.bitwarden-sm_secret.db_admin_secret.note}",
        "project_id" : "${resource.bitwarden-sm_secret.db_admin_secret.project_id}",
        "organization_id" : "${resource.bitwarden-sm_secret.db_admin_secret.organization_id}",
        "creation_date" : "${resource.bitwarden-sm_secret.db_admin_secret.creation_date}",
        "revision_date" : "${resource.bitwarden-sm_secret.db_admin_secret.revision_date}"
    }
}
EOF
  filename = "${path.module}/db_admin_credentials.json"
}

resource "local_file" "service_account_secret" {
  content  = <<EOF
{
    "db_username_secret": {
        "id" : "${resource.bitwarden-sm_secret.service_account_secret.id}",
        "key" : "${resource.bitwarden-sm_secret.service_account_secret.key}",
        "value" : "${resource.bitwarden-sm_secret.service_account_secret.value}",
        "note" : "${resource.bitwarden-sm_secret.service_account_secret.note}",
        "project_id" : "${resource.bitwarden-sm_secret.service_account_secret.project_id}",
        "organization_id" : "${resource.bitwarden-sm_secret.service_account_secret.organization_id}",
        "creation_date" : "${resource.bitwarden-sm_secret.service_account_secret.creation_date}",
        "revision_date" : "${resource.bitwarden-sm_secret.service_account_secret.revision_date}"
    }
}
EOF
  filename = "${path.module}/service_account_secret.json"
}
