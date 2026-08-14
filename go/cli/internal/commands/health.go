package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check API health",
	Run: func(cmd *cobra.Command, _ []string) {
		_, c := mustClient()
		resp, err := c.HealthWithResponse(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		fmt.Printf("Status: %s\n", resp.JSON200.Status)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
