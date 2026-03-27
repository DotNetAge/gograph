package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/DotNetAge/gograph/pkg/api"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui [db_path]",
	Short: "Start the interactive Terminal User Interface (TUI)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := args[0]
		return runTUI(dbPath)
	},
}

func runTUI(dbPath string) error {
	db, err := api.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	completer := readline.NewPrefixCompleter(
		readline.PcItem("/help"),
		readline.PcItem("/exit"),
		readline.PcItem("/quit"),
		readline.PcItem("/exec"),
		readline.PcItem("/query"),
		readline.PcItem("CREATE"),
		readline.PcItem("MATCH"),
		readline.PcItem("RETURN"),
		readline.PcItem("SET"),
		readline.PcItem("DELETE"),
		readline.PcItem("REMOVE"),
		readline.PcItem("WHERE"),
	)

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[36mgograph>\033[0m ",
		HistoryFile:     "/tmp/gograph_readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return fmt.Errorf("failed to init readline: %w", err)
	}
	defer l.Close()

	color.Cyan("Welcome to GoGraph TUI!")
	color.Cyan("Connected to database: %s", dbPath)
	color.Cyan("Type /help for usage instructions, or /exit to quit.\n")

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if handleInternalCommand(line, db) {
			continue
		}

		// Implicitly handle cypher
		// Auto-detect Exec vs Query based on the keyword
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "CREATE ") || strings.HasPrefix(upper, "SET ") ||
			strings.HasPrefix(upper, "DELETE ") || strings.HasPrefix(upper, "REMOVE ") {
			executeExec(db, line)
		} else {
			executeQuery(db, line)
		}
	}

	color.Cyan("Goodbye!")
	return nil
}

func handleInternalCommand(line string, db *api.DB) bool {
	if !strings.HasPrefix(line, "/") {
		return false
	}

	parts := strings.SplitN(line, " ", 2)
	cmd := parts[0]
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	switch cmd {
	case "/exit", "/quit":
		color.Cyan("Goodbye!")
		os.Exit(0)
	case "/help":
		printHelp()
	case "/exec":
		if args == "" {
			color.Red("Error: missing cypher query for /exec")
		} else {
			executeExec(db, args)
		}
	case "/query":
		if args == "" {
			color.Red("Error: missing cypher query for /query")
		} else {
			executeQuery(db, args)
		}
	default:
		color.Red("Unknown command: %s. Type /help for available commands.", cmd)
	}

	return true
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  /help           Show this help message")
	fmt.Println("  /exit, /quit    Exit the TUI")
	fmt.Println("  /exec <cypher>  Explicitly run a data modification query (CREATE, SET, DELETE, REMOVE)")
	fmt.Println("  /query <cypher> Explicitly run a data retrieval query (MATCH ... RETURN)")
	fmt.Println("\nOr simply type your Cypher query directly. The TUI will automatically route it:")
	fmt.Println("  - CREATE, SET, DELETE, REMOVE -> Executed as /exec")
	fmt.Println("  - MATCH -> Executed as /query")
}

func executeExec(db *api.DB, query string) {
	res, err := db.Exec(context.Background(), query)
	if err != nil {
		color.Red("Error: %v", err)
		return
	}
	color.Green("OK.")
	fmt.Printf("Nodes modified: %d | Relationships modified: %d\n", res.AffectedNodes, res.AffectedRels)
}

func executeQuery(db *api.DB, query string) {
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		color.Red("Error: %v", err)
		return
	}
	defer rows.Close()

	cols := rows.Columns()
	if len(cols) == 0 {
		fmt.Println("Empty result set.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(cols)
	table.SetAutoWrapText(false)

	count := 0
	for rows.Next() {
		count++
		scanArgs := make([]interface{}, len(cols))
		dest := make([]interface{}, len(cols))
		for i := range scanArgs {
			scanArgs[i] = &dest[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			color.Red("Scan error: %v", err)
			return
		}

		rowStrs := make([]string, len(cols))
		for i, v := range dest {
			rowStrs[i] = formatValue(v)
		}
		table.Append(rowStrs)
	}

	if count > 0 {
		table.Render()
		fmt.Printf("(%d rows)\n", count)
	} else {
		fmt.Println("(no rows)")
	}
}
