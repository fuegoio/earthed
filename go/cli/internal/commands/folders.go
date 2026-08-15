package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fuegoio/planetary/go/sdk/planetary"
	"github.com/spf13/cobra"
)

var foldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "Manage feed folders",
}

var foldersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all folders",
	Run: func(cmd *cobra.Command, _ []string) {
		_, c := mustClient()
		resp, err := c.ListFoldersWithResponse(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE")
		for _, fo := range *resp.JSON200 {
			fmt.Fprintf(w, "%d\t%s\n", fo.Id, fo.Title)
		}
		w.Flush()
	},
}

var foldersCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a folder",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		resp, err := c.CreateFolderWithResponse(context.Background(), planetary.CreateFolderJSONRequestBody{
			Title: args[0],
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		fo := *resp.JSON200
		fmt.Printf("Created folder %d: %s\n", fo.Id, fo.Title)
	},
}

var foldersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a folder",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		id, err := parseInt64(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid id: %v\n", err)
			os.Exit(1)
		}
		resp, err := c.DeleteFolderWithResponse(context.Background(), id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.StatusCode() != 204 {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		fmt.Printf("Deleted folder %d\n", id)
	},
}

func init() {
	foldersCmd.AddCommand(foldersListCmd, foldersCreateCmd, foldersDeleteCmd)
	rootCmd.AddCommand(foldersCmd)
}
