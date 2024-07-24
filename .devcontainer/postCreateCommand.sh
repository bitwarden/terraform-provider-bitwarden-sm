#!/bin/sh
sudo apt-get update
sudo apt-get install -y musl-tools

git config --global --add safe.directory "$PWD"

# configure Go to use musl-gcc
go env -w CC=musl-gcc

# create ~/.terraformrc dev override
cat <<EOF > ~/.terraformrc
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
  go install -ldflags '-linkmode external -extldflags \"-static -Wl,-unresolved-symbols=ignore-all\"' .
"
