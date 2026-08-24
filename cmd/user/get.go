package user

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
	Short: "Get an IBM Verify user by ID",
	RunE:  runGetUser,
}

var (
	getTenant       string
	getClientID     string
	getClientSecret string
	getUserID       string
)

func init() {
	UserCmd.AddCommand(getCmd)

	getCmd.Flags().StringVar(&getTenant, "tenant", "", "IBM Verify tenant URL (required)")
	getCmd.Flags().StringVar(&getClientID, "client-id", "", "Client ID (required)")
	getCmd.Flags().StringVar(&getClientSecret, "client-secret", "", "Client secret (required)")
	getCmd.Flags().StringVar(&getUserID, "id", "", "User ID (required)")

	_ = getCmd.MarkFlagRequired("tenant")
	_ = getCmd.MarkFlagRequired("client-id")
	_ = getCmd.MarkFlagRequired("client-secret")
	_ = getCmd.MarkFlagRequired("id")
}

func runGetUser(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(getTenant, client.WithClientCredentials(getClientID, getClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	u, err := cl.Users.Get(ctx, getUserID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, u)
	default:
		id, _ := u["id"].(string)
		userName, _ := u["userName"].(string)
		displayName, _ := u["displayName"].(string)
		active := fmt.Sprintf("%v", u["active"])
		fmt.Fprintf(c.OutOrStdout(), "ID:          %s\nUserName:    %s\nDisplayName: %s\nActive:      %s\n",
			id, userName, displayName, active)
	}
	return nil
}
