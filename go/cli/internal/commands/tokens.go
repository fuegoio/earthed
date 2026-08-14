package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fuegoio/planetary/go/sdk/planetary"
	"github.com/spf13/cobra"
)

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Manage API tokens",
}

var tokensListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API tokens",
	Run: func(cmd *cobra.Command, _ []string) {
		_, c := mustClient()
		resp, err := c.ListTokensWithResponse(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tLABEL\tCREATED\tLAST USED")
		for _, t := range *resp.JSON200 {
			lastUsed := "never"
			if t.LastUsedAt != nil {
				lastUsed = t.LastUsedAt.Format("2006-01-02 15:04")
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.Id, t.Label, t.CreatedAt.Format("2006-01-02 15:04"), lastUsed)
		}
		w.Flush()
	},
}

var tokensCreateCmd = &cobra.Command{
	Use:   "create <label>",
	Short: "Create an API token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		resp, err := c.CreateTokenWithResponse(context.Background(), planetary.CreateTokenJSONRequestBody{
			Label: args[0],
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		t := *resp.JSON200
		fmt.Printf("Token %d: %s\n", t.Id, t.Token)
		fmt.Println("Save this token — it won't be shown again.")
	},
}

var tokensDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an API token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		id, err := parseInt64(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid id: %v\n", err)
			os.Exit(1)
		}
		resp, err := c.DeleteTokenWithResponse(context.Background(), id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.StatusCode() != 204 {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		fmt.Printf("Deleted token %d\n", id)
	},
}

func init() {
	tokensCmd.AddCommand(tokensListCmd, tokensCreateCmd, tokensDeleteCmd)
	rootCmd.AddCommand(tokensCmd)
}
