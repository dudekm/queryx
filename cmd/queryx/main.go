package main

import (
	"context"
	ejson "encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dudekm/queryx"
)

var (
	gameType = flag.String("type", "minecraft", "Game server type (minecraft, cs16, etc.)")
	host     = flag.String("host", "", "Server hostname or IP")
	port     = flag.Int("port", 0, "Server port (0 = use default)")
	timeout  = flag.Duration("timeout", 5*time.Second, "Query timeout")
	debug    = flag.Bool("debug", false, "Enable debug logging")
	json     = flag.Bool("json", false, "Output as JSON")
	verbose  = flag.Bool("verbose", false, "Show detailed diagnostic information (DNS, SRV, timing)")
	version  = flag.Bool("version", false, "Show version information")
)

// formatPing formats ping in milliseconds
func formatPing(ms int) string {
	return fmt.Sprintf("%dms", ms)
}

const (
	appVersion = "1.0.0"
	appName    = "QueryX"
)

func main() {
	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("%s v%s\n", appName, appVersion)
		fmt.Println("Universal game server query tool")
		os.Exit(0)
	}

	// Validate required flags
	if *host == "" {
		fmt.Fprintf(os.Stderr, "Error: -host is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create client with options
	opts := []queryx.Option{
		queryx.WithTimeout(*timeout),
	}

	if *debug {
		opts = append(opts, queryx.WithDebug(true))
	}

	client := queryx.NewClientWithDefaults(opts...)

	// Prepare port pointer
	var portPtr *int
	if *port > 0 {
		portPtr = port
	}

	// Execute query
	ctx := context.Background()

	if *verbose {
		// Use verbose query mode
		verboseResult, err := client.QueryVerbose(ctx, queryx.GameType(*gameType), *host, portPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Output result with diagnostics
		if *json {
			printVerboseJSON(verboseResult)
		} else {
			printVerboseFormatted(verboseResult)
		}
	} else {
		// Use normal query mode
		result, err := client.Query(ctx, queryx.GameType(*gameType), *host, portPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Output result
		if *json {
			printJSON(result)
		} else {
			printFormatted(result)
		}
	}
}

func printFormatted(result *queryx.QueryResult) {
	fmt.Printf("═══════════════════════════════════════\n")
	fmt.Printf("  %s\n", result.Name)
	fmt.Printf("═══════════════════════════════════════\n")
	fmt.Printf("\n")

	if result.Online {
		fmt.Printf("  Status:       🟢 ONLINE\n")
	} else {
		fmt.Printf("  Status:       🔴 OFFLINE\n")
	}

	fmt.Printf("  Players:      %d/%d", result.NumPlayers, result.MaxPlayers)
	if result.Bots > 0 {
		fmt.Printf(" (%d bots)", result.Bots)
	}
	fmt.Println()

	if result.Map != "" {
		fmt.Printf("  Map:          %s\n", result.Map)
	}

	if result.Version != "" {
		fmt.Printf("  Version:      %s\n", result.Version)
	}

	fmt.Printf("  Type:         %s\n", result.Type)
	fmt.Printf("  Ping:         %s\n", formatPing(result.Ping))

	// Show player list if available
	if len(result.Players) > 0 {
		fmt.Printf("\n  Players online:\n")
		for i, player := range result.Players {
			if i >= 10 {
				fmt.Printf("    ... and %d more\n", len(result.Players)-10)
				break
			}
			fmt.Printf("    - %s", player.Name)
			if player.Score > 0 {
				fmt.Printf(" (score: %d)", player.Score)
			}
			if player.Duration > 0 {
				fmt.Printf(" [%v]", player.Duration)
			}
			fmt.Println()
		}
	}

	// Show protocol-specific data (Raw)
	if result.Raw != nil {
		fmt.Printf("\n  Protocol-specific data:\n")
		fmt.Printf("    Available in Raw field (use -json to see full data)\n")
	}

	fmt.Printf("\n")
}

func printVerboseFormatted(verboseResult *queryx.VerboseQueryResult) {
	result := verboseResult.Result
	diag := verboseResult.Diagnostics

	// Print standard result
	printFormatted(result)

	// Print diagnostic information
	fmt.Printf("═══════════════════════════════════════\n")
	fmt.Printf("  DIAGNOSTIC INFORMATION\n")
	fmt.Printf("═══════════════════════════════════════\n")
	fmt.Printf("\n")

	// Query metadata
	fmt.Printf("  Timestamp:    %s\n", diag.Timestamp.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("  Protocol:     %s\n", diag.QueryMetrics.Protocol)
	if diag.QueryMetrics.ProtocolVersion > 0 {
		fmt.Printf("  Protocol Ver: %d\n", diag.QueryMetrics.ProtocolVersion)
	}
	fmt.Printf("  Success:      %v\n", diag.QueryMetrics.Success)

	// Timing information
	fmt.Printf("\n  Timing:\n")
	fmt.Printf("    DNS Lookup:   %dms\n", diag.QueryMetrics.DNSLatencyMs)
	fmt.Printf("    Server Query: %dms\n", diag.QueryMetrics.QueryLatencyMs)
	fmt.Printf("    Network Ping: %dms\n", diag.QueryMetrics.LatencyMs)
	totalTime := diag.QueryMetrics.DNSLatencyMs + diag.QueryMetrics.QueryLatencyMs
	fmt.Printf("    Total Time:   %dms\n", totalTime)

	// DNS Resolution
	fmt.Printf("\n  DNS Resolution:\n")
	fmt.Printf("    Input Host:   %s\n", diag.Resolution.InputHostname)
	fmt.Printf("    Resolved IP:  %s\n", diag.Resolution.ResolvedIP)
	fmt.Printf("    Resolved Port: %d\n", diag.Resolution.ResolvedPort)

	// SRV Records
	if diag.Resolution.SRVRecordFound {
		fmt.Printf("\n  SRV Records:  ✓ Found\n")
		for i, srv := range diag.Resolution.SRVRecords {
			fmt.Printf("    [%d] %s:%d (priority: %d, weight: %d)\n",
				i+1, srv.Target, srv.Port, srv.Priority, srv.Weight)
		}
	} else {
		fmt.Printf("\n  SRV Records:  ✗ Not found\n")
	}

	// A/AAAA Records
	if len(diag.Resolution.ARecords) > 0 {
		fmt.Printf("\n  A Records (IPv4):\n")
		for _, ip := range diag.Resolution.ARecords {
			fmt.Printf("    - %s\n", ip)
		}
	}

	if len(diag.Resolution.AAAARecords) > 0 {
		fmt.Printf("\n  AAAA Records (IPv6):\n")
		for _, ip := range diag.Resolution.AAAARecords {
			fmt.Printf("    - %s\n", ip)
		}
	}

	fmt.Printf("\n")
}

func printJSON(result *queryx.QueryResult) {
	data, err := ejson.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printVerboseJSON(verboseResult *queryx.VerboseQueryResult) {
	data, err := ejson.MarshalIndent(verboseResult, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
