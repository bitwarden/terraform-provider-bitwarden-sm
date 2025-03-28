---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden-sm_projects Data Source - terraform-provider-bitwarden-sm"
subcategory: "Data Source"
description: |-
  The `projects` data source fetches all projects accessible by the used machine account.
---

# bitwarden-sm_projects (Data Source)

The `projects` data source fetches all projects accessible by the used machine account.

## Example usage

```terraform
data "bitwarden-sm_projects" "projects" {}

output "projects" {
  value = data.bitwarden-sm_projects.projects
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `projects` (Attributes List) Nested list of all fetched projects. (see [below for nested schema](#nestedatt--projects))

<a id="nestedatt--projects"></a>
### Nested Schema for `projects`

Read-Only:

- `creation_date` (String) String representation of the creation date of the project.
- `id` (String) String representation of the `ID` of the project inside Bitwarden Secrets Manager.
- `name` (String) String representation of the `name` of the secret inside Bitwarden Secrets Manager.
- `organization_id` (String) String representation of the `ID` of the organization to which the project belongs.
- `revision_date` (String) String representation of the revision date of the project.
