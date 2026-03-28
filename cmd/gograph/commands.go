package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DotNetAge/gograph/pkg/api"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [db_path] [cypher_query]",
	Short: "Execute a data modification Cypher query (CREATE, SET, DELETE, REMOVE)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := defaultDBPath
		query := ""
		if len(args) == 1 {
			query = args[0]
		} else {
			dbPath = args[0]
			query = args[1]
		}

		db, err := api.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		res, err := db.Exec(context.Background(), query)
		if err != nil {
			color.Red("Execution error: %v", err)
			return nil
		}

		color.Green("Query executed successfully.")
		fmt.Printf("Affected Nodes: %d\n", res.AffectedNodes)
		fmt.Printf("Affected Rels:  %d\n", res.AffectedRels)
		return nil
	},
}

var queryCmd = &cobra.Command{
	Use:   "query [db_path] [cypher_query]",
	Short: "Execute a data retrieval Cypher query (MATCH ... RETURN)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := defaultDBPath
		query := ""
		if len(args) == 1 {
			query = args[0]
		} else {
			dbPath = args[0]
			query = args[1]
		}

		db, err := api.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		rows, err := db.Query(context.Background(), query)
		if err != nil {
			color.Red("Query error: %v", err)
			return nil
		}
		defer rows.Close()

		cols := rows.Columns()
		if len(cols) == 0 {
			fmt.Println("Empty result set.")
			return nil
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
				return fmt.Errorf("scan error: %w", err)
			}

			rowStrs := make([]string, len(cols))
			for i, v := range dest {
				rowStrs[i] = formatValue(v)
			}
			table.Append(rowStrs)
		}

		if count > 0 {
			table.Render()
			fmt.Printf("\nReturned %d rows.\n", count)
		} else {
			fmt.Println("No rows returned.")
		}

		return nil
	},
}

func formatValue(val interface{}) string {
	switch v := val.(type) {
	case *graph.Node:
		props := []string{}
		for k, pv := range v.Properties {
			var propVal string
			switch pv.Type() {
			case graph.PropertyTypeString:
				propVal = pv.StringValue()
			case graph.PropertyTypeInt:
				propVal = fmt.Sprintf("%d", pv.IntValue())
			case graph.PropertyTypeFloat:
				propVal = fmt.Sprintf("%f", pv.FloatValue())
			case graph.PropertyTypeBool:
				propVal = fmt.Sprintf("%t", pv.BoolValue())
			default:
				propVal = "unknown"
			}
			props = append(props, fmt.Sprintf("%s:%s", k, propVal))
		}
		labels := strings.Join(v.Labels, ":")
		if labels != "" {
			labels = ":" + labels
		}
		propStr := ""
		if len(props) > 0 {
			propStr = fmt.Sprintf(" {%s}", strings.Join(props, ", "))
		}
		return fmt.Sprintf("(%s%s%s)", v.ID, labels, propStr)
	case *graph.Relationship:
		props := []string{}
		for k, pv := range v.Properties {
			var propVal string
			switch pv.Type() {
			case graph.PropertyTypeString:
				propVal = pv.StringValue()
			case graph.PropertyTypeInt:
				propVal = fmt.Sprintf("%d", pv.IntValue())
			case graph.PropertyTypeFloat:
				propVal = fmt.Sprintf("%f", pv.FloatValue())
			case graph.PropertyTypeBool:
				propVal = fmt.Sprintf("%t", pv.BoolValue())
			default:
				propVal = "unknown"
			}
			props = append(props, fmt.Sprintf("%s:%s", k, propVal))
		}
		propStr := ""
		if len(props) > 0 {
			propStr = fmt.Sprintf(" {%s}", strings.Join(props, ", "))
		}
		return fmt.Sprintf("[%s:%s%s]", v.ID, v.Type, propStr)
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", v)
	}
}
