# Configuration-based authentication
provider "bitwarden-sm" {
    api_url         = "https://api.bitwarden.com"
    identity_url    = "https://identity.bitwarden.com"
    access_token    = "< secret machine account access token >"
    organization_id = "< your organization id >"
}
