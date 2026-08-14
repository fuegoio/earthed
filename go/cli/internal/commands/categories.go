package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fuegoio/planetary/go/sdk/planetary"
	"github.com/spf13/cobra"
)

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Manage feed categories",
}

var categoriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	Run: func(cmd *cobra.Command, _ []string) {
		_, c := mustClient()
		resp, err := c.ListCategoriesWithResponse(context.Background())
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
		for _, cat := range *resp.JSON200 {
			fmt.Fprintf(w, "%d\t%s\n", cat.Id, cat.Title)
		}
		w.Flush()
	},
}

var categoriesCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a category",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		resp, err := c.CreateCategoryWithResponse(context.Background(), planetary.CreateCategoryJSONRequestBody{
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
		cat := *resp.JSON200
		fmt.Printf("Created category %d: %s\n", cat.Id, cat.Title)
	},
}

var categoriesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a category",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		id, err := parseInt64(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid id: %v\n", err)
			os.Exit(1)
		}
		resp, err := c.DeleteCategoryWithResponse(context.Background(), id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.StatusCode() != 204 {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		fmt.Printf("Deleted category %d\n", id)
	},
}

func init() {
	categoriesCmd.AddCommand(categoriesListCmd, categoriesCreateCmd, categoriesDeleteCmd)
	rootCmd.AddCommand(categoriesCmd)
}
