data "bitwarden-sm_projects" "projects" {}

output "projects" {
  value = data.bitwarden-sm_projects.projects
}
