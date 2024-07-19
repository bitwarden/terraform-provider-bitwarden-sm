variable "api_url" {
  description = "URI for Bitwarden Secrets Manager API endpoint. May also be provided via BW_API_URL environment variable."
  type        = string

}

variable "identity_url" {
  description = "URI for Bitwarden Secrets Manager IDENTITY endpoint. May also be provided via BW_IDENTITY_API_URL environment variable."
  type        = string
}

variable "organization_id" {
  description = "Organization ID for Bitwarden Secrets Manager endpoints. May also be provided via BW_ORGANIZATION_ID environment variable."
  type        = string
}

variable "access_token" {
  description = "Access token for Bitwarden Secrets Manager endpoints. May also be provided via BW_ACCESS_TOKEN environment variable."
  type        = string
}

variable "key" {
  description = "The key of a secret in Bitwarden Secrets Manager."
  type        = string
}

variable "value" {
  description = "The value of a secret in Bitwarden Secrets Manager."
  type        = string
}

variable "note" {
  description = "The note of a secret in Bitwarden Secrets Manager."
  type        = string
}

variable "project_id" {
  description = "The Project ID of the secret in Bitwarden Secrets Manager."
  type        = string
}
