package main

import (
	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/cmd/app"
	"github.com/ChrisVerde02/ibmverify-cli/cmd/apiclient"
	"github.com/ChrisVerde02/ibmverify-cli/cmd/cert"
	"github.com/ChrisVerde02/ibmverify-cli/cmd/config"
	"github.com/ChrisVerde02/ibmverify-cli/cmd/token"
	"github.com/ChrisVerde02/ibmverify-cli/cmd/user"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=v1.7.0" ./cmd/ibmverify
var version = "dev"

func main() {
	cmd.Root().AddCommand(app.AppCmd)
	cmd.Root().AddCommand(apiclient.APIClientCmd)
	cmd.Root().AddCommand(cert.CertCmd)
	cmd.Root().AddCommand(token.TokenCmd)
	cmd.Root().AddCommand(config.ConfigCmd)
	cmd.Root().AddCommand(user.UserCmd)
	cmd.Execute(version)
}
