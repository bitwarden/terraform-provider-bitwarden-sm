terraform {
  required_providers {
    bitwarden-secrets = {
      source = "registry.terraform.io/bitwarden/bitwarden-secrets"
    }
  }
}

# Create a project in Bitwarden Secrets Manager
resource "bitwarden-secrets_project" "example" {
  name = "terraform-example-project"
}

# Output the project details
output "project" {
  value = {
    id              = bitwarden-secrets_project.example.id
    name            = bitwarden-secrets_project.example.name
    organization_id = bitwarden-secrets_project.example.organization_id
    creation_date   = bitwarden-secrets_project.example.creation_date
    revision_date   = bitwarden-secrets_project.example.revision_date
  }
}

