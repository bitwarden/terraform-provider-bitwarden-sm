---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden-sm_secret Data Source - terraform-provider-bitwarden-sm"
subcategory: "Data Source"
description: |-
  The `secret` data source fetches a particular secret from Bitwarden Secrets Manager based on a given `ID`.
---

# bitwarden-sm_secret (Data Source)

The `secret` data source fetches a particular secret from Bitwarden Secrets Manager based on a given `ID`.

## Example usage

```terraform
data "bitwarden-sm_secret" "secret" {
  id = "e6a8066c-81e6-428e-bf5d-b1b900fe1b42"
}

output "secret" {
  value = {
    id  = data.bitwarden-sm_secret.secret.id
    key = data.bitwarden-sm_secret.secret.key
    # The actual secret value is marked sensitive and will not be printed to stdout
    # value          = data.bitwarden-sm_secret.secret.value
    note            = resource.bitwarden-sm_secret.secret.note
    project_id      = resource.bitwarden-sm_secret.secret.project_id
    organization_id = resource.bitwarden-sm_secret.secret.organization_id
    creation_date   = resource.bitwarden-sm_secret.secret.creation_date
    revision_date   = resource.bitwarden-sm_secret.secret.revision_date
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) String representation of the `ID` of the secret inside Bitwarden Secrets Manager.

### Read-Only

- `creation_date` (String) String representation of the creation date of the secret.
- `key` (String) String representation of the `key` of the secret. Inside Bitwarden Secrets Manager this is called "name".
- `note` (String) String representation of the `note` of the secret inside Bitwarden Secrets Manager.
- `organization_id` (String) String representation of the `ID` of the organization to which the secret belongs.
- `project_id` (String) String representation of the `ID` of the project to which the secret belongs. If the used machine account has no read access to this project, access will not be granted.
- `revision_date` (String) String representation of the revision date of the secret.
- `value` (String, Sensitive) String representation of the `value` of the secret inside Bitwarden Secrets Manager. This attribute is sensitive.
