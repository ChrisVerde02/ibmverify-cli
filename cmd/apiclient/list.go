package apiclient

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List IBM Verify API clients",
	RunE:  runListAPIClients,
}

var (
	listTenant       string
	listClientID     string
	listClientSecret string
)

func init() {
	APIClientCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listTenant, "tenant", "", "IBM Verify tenant URL (required)")
	listCmd.Flags().StringVar(&listClientID, "client-id", "", "Client ID (required)")
	listCmd.Flags().StringVar(&listClientSecret, "client-secret", "", "Client secret (required)")

	_ = listCmd.MarkFlagRequired("tenant")
	_ = listCmd.MarkFlagRequired("client-id")
	_ = listCmd.MarkFlagRequired("client-secret")
}

func runListAPIClients(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(listTenant, client.WithClientCredentials(listClientID, listClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	list, err := cl.APIClients.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("list api clients: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, list)
	default:
		if len(list) == 0 {
			fmt.Fprintln(c.OutOrStdout(), "No API clients found.")
			return nil
		}
		for _, ac := range list {
			id, _ := ac["clientId"].(string)
			name, _ := ac["clientName"].(string)
			enabled := fmt.Sprintf("%v", ac["enabled"])
			fmt.Fprintf(c.OutOrStdout(), "%-36s  %-40s  %s\n", id, name, enabled)
		}
	}
	return nil
}
