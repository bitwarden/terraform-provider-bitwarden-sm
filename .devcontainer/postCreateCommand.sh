#!/bin/sh
git config --global --add safe.directory "$PWD"

# create ~/.terraformrc dev override
cat <<EOF >~/.terraformrc
provider_installation {
  dev_overrides {
      "registry.terraform.io/bitwarden/bitwarden-sm" = "/go/bin"
  }
  direct {}
}
EOF

echo "
devcontainer setup complete!

To build and install the terraform provider, run:
  go install -ldflags '-linkmode external -extldflags \"-lm\"' .
"
