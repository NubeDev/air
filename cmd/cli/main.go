package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	apiclient "github.com/NubeDev/air/clients/go"
	"github.com/spf13/cobra"
)

var (
	serverURL    = flag.String("server", "http://localhost:9000", "AIR server URL")
	authToken    = flag.String("token", "", "JWT authentication token")
	authDisabled = flag.Bool("auth", false, "Disable authentication")
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "aircli",
		Short: "AIR CLI - AI Reporter command line interface",
		Long:  `AIR CLI provides command-line access to the AIR (AI Reporter) system for managing datasources, reports, and analytics.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(serverURL, "server", "http://localhost:9000", "AIR server URL")
	rootCmd.PersistentFlags().StringVar(authToken, "token", "", "JWT authentication token")
	rootCmd.PersistentFlags().BoolVar(authDisabled, "auth", false, "Disable authentication")

	// Datasource commands
	datasourceCmd := &cobra.Command{
		Use:   "datasource",
		Short: "Manage datasources",
		Long:  `Manage analytics datasource connections.`,
	}
	datasourceCmd.AddCommand(createDatasourceCmd())
	datasourceCmd.AddCommand(listDatasourcesCmd())
	datasourceCmd.AddCommand(healthCheckCmd())
	rootCmd.AddCommand(datasourceCmd)

	// Learn commands
	rootCmd.AddCommand(learnCmd())

	// Report commands
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Manage reports",
		Long:  `Manage report definitions and execution.`,
	}
	reportCmd.AddCommand(createReportCmd())
	reportCmd.AddCommand(listReportsCmd())
	reportCmd.AddCommand(runReportCmd())
	rootCmd.AddCommand(reportCmd)

	// Generic HTTP commands
	rootCmd.AddCommand(createGenericCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func createDatasourceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new datasource",
		Long:  `Create a new analytics datasource connection.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Datasource creation not implemented yet")
		},
	}
}

func listDatasourcesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all datasources",
		Long:  `List all registered analytics datasources with their health status.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create API client
			client, err := apiclient.NewClientWithResponses(*serverURL)
			if err != nil {
				log.Fatalf("Failed to create API client: %v", err)
			}

			// Make request
			ctx := context.Background()
			resp, err := client.GetV1DatasourcesWithResponse(ctx)
			if err != nil {
				log.Fatalf("Failed to get datasources: %v", err)
			}

			// Check response status
			if resp.StatusCode() != 200 {
				log.Fatalf("API request failed with status %d: %s", resp.StatusCode(), resp.Body)
			}

			// Parse response
			if resp.JSON200 == nil {
				log.Fatalf("Failed to parse response: no JSON data")
			}

			// Display results
			datasources := resp.JSON200.Datasources
			if datasources == nil {
				datasources = &[]apiclient.DatasourceResponse{}
			}
			fmt.Printf("Found %d datasources:\n", len(*datasources))
			for _, ds := range *datasources {
				status := "❌"
				if ds.HealthStatus != nil && *ds.HealthStatus == "healthy" {
					status = "✅"
				}
				fmt.Printf("  %s %s (%s) - %s\n", status, *ds.Id, *ds.Kind, *ds.DisplayName)
				if ds.Error != nil && *ds.Error != "" {
					fmt.Printf("    Error: %s\n", *ds.Error)
				}
			}
		},
	}
}

func healthCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health [id]",
		Short: "Check datasource health",
		Long:  `Check the health status of a specific datasource.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Health check for datasource %s not implemented yet\n", args[0])
		},
	}
}

func learnCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "learn [datasource_id]",
		Short: "Learn database schema",
		Long:  `Introspect a datasource and learn its schema structure.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Learning from datasource %s not implemented yet\n", args[0])
		},
	}
}

func createReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new report",
		Long:  `Create a new report definition.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Report creation not implemented yet")
		},
	}
}

func listReportsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all reports",
		Long:  `List all saved reports.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Report listing not implemented yet")
		},
	}
}

func runReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [key]",
		Short: "Run a report",
		Long:  `Execute a saved report with parameters.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Running report %s not implemented yet\n", args[0])
		},
	}
}

func createGenericCmd() *cobra.Command {
	var method string
	var path string
	var jsonData string
	var queryParams []string

	cmd := &cobra.Command{
		Use:   "http [method] [path]",
		Short: "Make generic HTTP requests",
		Long:  `Make generic HTTP requests to the AIR API.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			method = args[0]
			path = args[1]
			fmt.Printf("Generic HTTP %s %s not implemented yet\n", method, path)
			if jsonData != "" {
				fmt.Printf("JSON data: %s\n", jsonData)
			}
			if len(queryParams) > 0 {
				fmt.Printf("Query params: %v\n", queryParams)
			}
		},
	}

	cmd.Flags().StringVar(&jsonData, "json", "", "JSON data to send in request body")
	cmd.Flags().StringArrayVar(&queryParams, "query", []string{}, "Query parameters (key=value)")

	return cmd
}
