package main

import (
	"context"
	"flag"
	"github.com/bitwarden/terraform-provider-bitwarden-sm/internal/provider"
	"github.com/bitwarden/terraform-provider-bitwarden-sm/version"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but it is suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/bitwarden/bitwarden-sm",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version.ProviderVersion), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
