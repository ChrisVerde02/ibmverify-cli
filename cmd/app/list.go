package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all IBM Verify applications",
	RunE:  runListApps,
}

var (
	listTenant       string
	listClientID     string
	listClientSecret string
)

func init() {
	AppCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listTenant, "tenant", "", "IBM Verify tenant URL (required)")
	listCmd.Flags().StringVar(&listClientID, "client-id", "", "Client ID (required)")
	listCmd.Flags().StringVar(&listClientSecret, "client-secret", "", "Client secret (required)")

	_ = listCmd.MarkFlagRequired("tenant")
	_ = listCmd.MarkFlagRequired("client-id")
	_ = listCmd.MarkFlagRequired("client-secret")
}

func runListApps(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(listTenant, client.WithClientCredentials(listClientID, listClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	list, err := cl.Apps.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("list apps: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, list)
	default:
		if len(list) == 0 {
			fmt.Fprintln(c.OutOrStdout(), "No applications found.")
			return nil
		}
		for _, a := range list {
			id, _ := a["id"].(string)
			if id == "" {
				id, _ = a["applicationID"].(string)
			}
			name, _ := a["name"].(string)
			state := fmtField(a, "applicationState")
			fmt.Fprintf(c.OutOrStdout(), "%-36s  %-40s  %s\n", id, name, state)
		}
	}
	return nil
}

// fmtField returns a field from a map as a string regardless of underlying type.
func fmtField(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}
