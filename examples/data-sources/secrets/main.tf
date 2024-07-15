data "bitwarden-sm_secrets" "secrets" {}

output "secrets" {
    value = data.bitwarden-sm_secrets.secrets
}
