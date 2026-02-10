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
	version  = flag.Bool("version", false, "Show version information")
)

// formatPing formats ping in milliseconds with appropriate precision
func formatPing(d time.Duration) string {
	ms := float64(d.Microseconds()) / 1000.0

	if ms < 1 {
		// Sub-millisecond: show 2 decimal places (e.g., "0.85ms")
		return fmt.Sprintf("%.2fms", ms)
	} else if ms < 10 {
		// 1-10ms: show 2 decimal places (e.g., "5.23ms")
		return fmt.Sprintf("%.2fms", ms)
	} else if ms < 100 {
		// 10-100ms: show 1 decimal place (e.g., "53.8ms")
		return fmt.Sprintf("%.1fms", ms)
	} else {
		// >= 100ms: show whole number (e.g., "582ms")
		return fmt.Sprintf("%.0fms", ms)
	}
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
	result, err := client.Query(ctx, queryx.GameType(*gameType), *host, portPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Output result
	if *json {
		printJSON(result, *debug)
	} else {
		printFormatted(result, *debug)
	}
}

func printFormatted(result *queryx.QueryResult, debug bool) {
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

func printJSON(result *queryx.QueryResult, debug bool) {
	data, err := ejson.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
