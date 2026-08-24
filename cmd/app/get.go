package app

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an IBM Verify application by ID",
	RunE:  runGetApp,
}

var (
	getTenant       string
	getClientID     string
	getClientSecret string
	getAppID        string
)

func init() {
	AppCmd.AddCommand(getCmd)

	getCmd.Flags().StringVar(&getTenant, "tenant", "", "IBM Verify tenant URL (required)")
	getCmd.Flags().StringVar(&getClientID, "client-id", "", "Client ID (required)")
	getCmd.Flags().StringVar(&getClientSecret, "client-secret", "", "Client secret (required)")
	getCmd.Flags().StringVar(&getAppID, "id", "", "Application ID (required)")

	_ = getCmd.MarkFlagRequired("tenant")
	_ = getCmd.MarkFlagRequired("client-id")
	_ = getCmd.MarkFlagRequired("client-secret")
	_ = getCmd.MarkFlagRequired("id")
}

func runGetApp(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(getTenant, client.WithClientCredentials(getClientID, getClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	a, err := cl.Apps.Get(ctx, getAppID)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, a)
	default:
		id, _ := a["id"].(string)
		if id == "" {
			id, _ = a["applicationID"].(string)
		}
		name, _ := a["name"].(string)
		state := fmtField(a, "applicationState")
		fmt.Fprintf(c.OutOrStdout(), "ID:    %s\nName:  %s\nState: %s\n", id, name, state)
	}
	return nil
}
