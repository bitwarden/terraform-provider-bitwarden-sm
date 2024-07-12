data "bitwarden-sm_secrets" "secrets" {
    organization_id = "< your organization id >"
}

output "secrets" {
    value = data.bitwarden-sm_secrets.secrets
}
