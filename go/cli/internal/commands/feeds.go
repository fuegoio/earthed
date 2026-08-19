package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fuegoio/earthed/go/sdk/earthed"
	"github.com/spf13/cobra"
)

var feedsCmd = &cobra.Command{
	Use:   "feeds",
	Short: "Manage feed subscriptions",
}

var feedsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all feeds",
	Run: func(cmd *cobra.Command, _ []string) {
		_, c := mustClient()
		resp, err := c.ListFeedsWithResponse(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tURL\tSITE")
		for _, f := range *resp.JSON200 {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", f.Id, f.Title, f.FeedUrl, f.SiteUrl)
		}
		w.Flush()
	},
}

var feedsAddCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Subscribe to a feed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		resp, err := c.CreateFeedWithResponse(context.Background(), earthed.CreateFeedJSONRequestBody{
			FeedUrl: args[0],
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		f := *resp.JSON200
		fmt.Printf("Subscribed to feed %d: %s\n", f.Id, f.Title)
	},
}

var feedsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Unsubscribe from a feed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		id, err := parseInt64(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid id: %v\n", err)
			os.Exit(1)
		}
		resp, err := c.DeleteFeedWithResponse(context.Background(), id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.StatusCode() != 204 {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		fmt.Printf("Deleted feed %d\n", id)
	},
}

func init() {
	feedsCmd.AddCommand(feedsListCmd, feedsAddCmd, feedsDeleteCmd)
	rootCmd.AddCommand(feedsCmd)
}
