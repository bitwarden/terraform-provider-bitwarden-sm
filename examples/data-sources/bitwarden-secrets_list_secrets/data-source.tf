# List all secrets accessible by the machine account
data "bitwarden-secrets_list_secrets" "all_secrets" {}

output "all_secrets" {
  value = data.bitwarden-secrets_list_secrets.all_secrets
}

# Filter secrets by project ID
data "bitwarden-secrets_list_secrets" "project_secrets" {
  project_id = "00000000-0000-0000-0000-000000000000"
}

output "project_secrets" {
  value = data.bitwarden-secrets_list_secrets.project_secrets
}

# Filter secrets by key (name) - case-insensitive substring match
data "bitwarden-secrets_list_secrets" "filtered_secrets" {
  filter = "database"
}

output "filtered_secrets" {
  value = data.bitwarden-secrets_list_secrets.filtered_secrets
}

# Combine project ID and key filters
data "bitwarden-secrets_list_secrets" "combined_filter" {
  project_id = "00000000-0000-0000-0000-000000000000"
  filter = "api-key"
}

output "combined_filter" {
  value = data.bitwarden-secrets_list_secrets.combined_filter
}
