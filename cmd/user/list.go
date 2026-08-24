package user

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	generated "github.com/ChrisVerde02/ibmverify-go/generated"
	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List IBM Verify users (SCIM v2)",
	RunE:  runListUsers,
}

var (
	listTenant       string
	listClientID     string
	listClientSecret string
	listFilter       string
)

func init() {
	UserCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listTenant, "tenant", "", "IBM Verify tenant URL (required)")
	listCmd.Flags().StringVar(&listClientID, "client-id", "", "Client ID (required)")
	listCmd.Flags().StringVar(&listClientSecret, "client-secret", "", "Client secret (required)")
	listCmd.Flags().StringVar(&listFilter, "filter", "", `SCIM filter expression, e.g. 'userName eq "john"'`)

	_ = listCmd.MarkFlagRequired("tenant")
	_ = listCmd.MarkFlagRequired("client-id")
	_ = listCmd.MarkFlagRequired("client-secret")
}

func runListUsers(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(listTenant, client.WithClientCredentials(listClientID, listClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	var req *generated.GetUsersRequest
	if listFilter != "" {
		req = &generated.GetUsersRequest{Filter: &listFilter}
	}

	list, err := cl.Users.List(ctx, req)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, list)
	default:
		if len(list) == 0 {
			fmt.Fprintln(c.OutOrStdout(), "No users found.")
			return nil
		}
		for _, u := range list {
			id, _ := u["id"].(string)
			userName, _ := u["userName"].(string)
			displayName, _ := u["displayName"].(string)
			fmt.Fprintf(c.OutOrStdout(), "%-36s  %-30s  %s\n", id, userName, displayName)
		}
	}
	return nil
}
