#!/bin/sh
git config --global --add safe.directory "$PWD"

sudo apt-get update && sudo apt-get install -y musl-tools

go env -w CGO_ENABLED="1"
go env -w CC="musl-gcc"
go env -w CGO_LDFLAGS="-static -Wl,-unresolved-symbols=ignore-all"

# create ~/.terraformrc dev override
cat <<EOF >~/.terraformrc
provider_installation {
  dev_overrides {
      "registry.terraform.io/bitwarden/bitwarden-sm" = "/go/bin"
      "registry.opentofu.org/bitwarden/bitwarden-sm" = "/go/bin"
  }
  direct {}
}
EOF

# install the provider
go install .

echo "
devcontainer setup complete!

To build a statically-linked binary, run:
  go env -w CGO_LDFLAGS=\"-static -Wl,-unresolved-symbols=ignore-all\"
  go env -w CC=\"musl-gcc\"
  go install .
"
