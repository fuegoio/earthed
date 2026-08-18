package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fuegoio/planetary/go/sdk/planetary"
	"github.com/spf13/cobra"
)

var entriesCmd = &cobra.Command{
	Use:   "entries",
	Short: "Browse feed entries",
}

var entriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List unread entries",
	Run: func(cmd *cobra.Command, _ []string) {
		_, c := mustClient()
		limit := int64(50)
		status := planetary.ListEntriesParamsStatusUnread
		resp, err := c.ListEntriesWithResponse(context.Background(), &planetary.ListEntriesParams{
			Limit:  &limit,
			Status: &status,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}

		// Build a feed-name lookup from the feeds endpoint.
		feedNames := map[int64]string{}
		if fr, ferr := c.ListFeedsWithResponse(context.Background()); ferr == nil && fr.JSON200 != nil {
			for _, f := range *fr.JSON200 {
				feedNames[f.Id] = f.Title
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSTARRED\tSTATUS\tDATE\tFEED\tTITLE")
		for _, e := range *resp.JSON200 {
			star := " "
			if e.Starred {
				star = "★"
			}
			status := e.Status
			date := e.PublishedAt.Format("2006-01-02")
			feed := feedNames[e.FeedId]
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", e.Id, star, status, date, feed, e.Title)
		}
		w.Flush()
	},
}

var entriesReadCmd = &cobra.Command{
	Use:   "read <id>",
	Short: "Read a single entry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, c := mustClient()
		id, err := parseInt64(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid id: %v\n", err)
			os.Exit(1)
		}
		resp, err := c.GetEntryWithResponse(context.Background(), id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resp.JSON200 == nil {
			printError(resp.HTTPResponse.StatusCode, resp.ApplicationproblemJSONDefault)
			os.Exit(1)
		}
		e := *resp.JSON200
		fmt.Printf("Title: %s\n", e.Title)
		if e.Author != nil {
			fmt.Printf("Author: %s\n", *e.Author)
		}
		fmt.Printf("URL: %s\n", e.Url)
		fmt.Printf("Published: %s\n", e.PublishedAt.Format("2006-01-02 15:04"))
		fmt.Printf("Status: %s\n\n", e.Status)
		if e.Description != nil {
			fmt.Println(*e.Description)
		}
	},
}

func init() {
	entriesCmd.AddCommand(entriesListCmd, entriesReadCmd)
	rootCmd.AddCommand(entriesCmd)
}
