// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"log"

	"terraform-provider-wordpress/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(
		context.Background(),
		provider.New("dev"),
		providerserver.ServeOpts{
			Address: "registry.terraform.io/avishayil/wordpress",
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
