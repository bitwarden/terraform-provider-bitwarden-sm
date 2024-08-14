data "bitwarden-sm_list_secrets" "secrets" {}

output "secrets" {
  value = data.bitwarden-sm_list_secrets.secrets
}
