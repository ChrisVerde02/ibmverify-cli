package user

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/ChrisVerde02/ibmverify-go/generated"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new IBM Verify user",
	RunE:  runCreateUser,
}

var (
	createTenant       string
	createClientID     string
	createClientSecret string
	createUserName     string
	createPassword     string
	createDisplayName  string
)

func init() {
	UserCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&createTenant, "tenant", "", "IBM Verify tenant URL (required)")
	createCmd.Flags().StringVar(&createClientID, "client-id", "", "Client ID (required)")
	createCmd.Flags().StringVar(&createClientSecret, "client-secret", "", "Client secret (required)")
	createCmd.Flags().StringVar(&createUserName, "username", "", "IBM Verify username (required)")
	createCmd.Flags().StringVar(&createPassword, "password", "", "Initial password (required)")
	createCmd.Flags().StringVar(&createDisplayName, "display-name", "", "Display name (optional)")

	_ = createCmd.MarkFlagRequired("tenant")
	_ = createCmd.MarkFlagRequired("client-id")
	_ = createCmd.MarkFlagRequired("client-secret")
	_ = createCmd.MarkFlagRequired("username")
	_ = createCmd.MarkFlagRequired("password")
}

func runCreateUser(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(createTenant, client.WithClientCredentials(createClientID, createClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	body := &generated.UserV2{
		Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		UserName: createUserName,
		Password: &createPassword,
	}
	if createDisplayName != "" {
		body.DisplayName = &createDisplayName
	}

	result, err := cl.Users.Create(ctx, &generated.CreateUserRequest{Body: body})
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, result)
	default:
		id, _ := result["id"].(string)
		fmt.Fprintf(c.OutOrStdout(), "✓ User created (userName=%s, id=%s)\n", createUserName, id)
	}
	return nil
}
