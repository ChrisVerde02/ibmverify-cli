package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	generated "github.com/ChrisVerde02/ibmverify-go/generated"
	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new IBM Verify application",
	RunE:  runCreateApp,
}

var (
	createTenant       string
	createClientID     string
	createClientSecret string
	createName         string
	createTemplateID   string
)

func init() {
	AppCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&createTenant, "tenant", "", "IBM Verify tenant URL (required)")
	createCmd.Flags().StringVar(&createClientID, "client-id", "", "Client ID (required)")
	createCmd.Flags().StringVar(&createClientSecret, "client-secret", "", "Client secret (required)")
	createCmd.Flags().StringVar(&createName, "name", "", "Application name (required)")
	createCmd.Flags().StringVar(&createTemplateID, "template-id", "", "Template ID (required)")

	_ = createCmd.MarkFlagRequired("tenant")
	_ = createCmd.MarkFlagRequired("client-id")
	_ = createCmd.MarkFlagRequired("client-secret")
	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("template-id")
}

func runCreateApp(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(createTenant, client.WithClientCredentials(createClientID, createClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	result, err := cl.Apps.Create(ctx, &generated.ApplicationRequestBean{
		Name:       createName,
		TemplateID: createTemplateID,
	})
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, result)
	default:
		// PostApplicationResponseBean only has _links; marshal to get the href
		b, _ := json.Marshal(result)
		var m map[string]interface{}
		_ = json.Unmarshal(b, &m)
		fmt.Fprintf(c.OutOrStdout(), "✓ Application created (name=%s)\n", createName)
	}
	return nil
}
