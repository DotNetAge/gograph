// Package format provides result formatters for gograph query output.
// It supports multiple output formats including JSON, XML, YAML, and CSV.
//
// Basic Usage:
//
//	formatter := format.NewFormatter("json")
//	output, err := formatter.Format(result)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(output)
package format

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/DotNetAge/gograph/pkg/cypher"
)

// Formatter defines the interface for formatting query results.
type Formatter interface {
	// Format converts a cypher.Result into a formatted string.
	Format(result cypher.Result) (string, error)

	// ContentType returns the MIME type for this formatter.
	ContentType() string
}

// NewFormatter creates a formatter for the given format name.
// Supported formats: "json", "xml", "yaml", "csv", "table" (default).
//
// Example:
//
//	formatter := format.NewFormatter("json")
//	output, err := formatter.Format(result)
func NewFormatter(name string) Formatter {
	switch strings.ToLower(name) {
	case "json":
		return &JSONFormatter{}
	case "xml":
		return &XMLFormatter{}
	case "yaml":
		return &YAMLFormatter{}
	case "csv":
		return &CSVFormatter{}
	default:
		return &TableFormatter{}
	}
}

// JSONFormatter formats results as JSON.
type JSONFormatter struct{}

// ContentType returns the MIME type for JSON.
func (f *JSONFormatter) ContentType() string {
	return "application/json"
}

// Format converts the result to a JSON string.
func (f *JSONFormatter) Format(result cypher.Result) (string, error) {
	type output struct {
		Columns       []string                 `json:"columns,omitempty"`
		Rows          []map[string]interface{} `json:"rows,omitempty"`
		AffectedNodes int                      `json:"affected_nodes,omitempty"`
		AffectedRels  int                      `json:"affected_rels,omitempty"`
		NodesCreated  int                      `json:"nodes_created,omitempty"`
		RelsCreated   int                      `json:"rels_created,omitempty"`
	}

	out := output{
		Columns:       result.Columns,
		Rows:          result.Rows,
		AffectedNodes: result.AffectedNodes,
		AffectedRels:  result.AffectedRels,
		NodesCreated:  result.NodesCreated,
		RelsCreated:   result.RelsCreated,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json format error: %w", err)
	}
	return string(data), nil
}

// XMLFormatter formats results as XML.
type XMLFormatter struct{}

// ContentType returns the MIME type for XML.
func (f *XMLFormatter) ContentType() string {
	return "application/xml"
}

// Format converts the result to an XML string.
func (f *XMLFormatter) Format(result cypher.Result) (string, error) {
	type xmlCell struct {
		Column string `xml:"column,attr"`
		Value  string `xml:",chardata"`
	}
	type xmlRow struct {
		Cells []xmlCell `xml:"cell"`
	}
	type xmlResult struct {
		XMLName       xml.Name `xml:"result"`
		Columns       []string `xml:"columns>column"`
		AffectedNodes int      `xml:"affected_nodes,omitempty"`
		AffectedRels  int      `xml:"affected_rels,omitempty"`
		NodesCreated  int      `xml:"nodes_created,omitempty"`
		RelsCreated   int      `xml:"rels_created,omitempty"`
		Rows          []xmlRow `xml:"rows>row"`
	}

	out := xmlResult{
		Columns:       result.Columns,
		AffectedNodes: result.AffectedNodes,
		AffectedRels:  result.AffectedRels,
		NodesCreated:  result.NodesCreated,
		RelsCreated:   result.RelsCreated,
	}

	for _, row := range result.Rows {
		xr := xmlRow{}
		for _, col := range result.Columns {
			val := ""
			if v, ok := row[col]; ok {
				val = fmt.Sprintf("%v", v)
			}
			xr.Cells = append(xr.Cells, xmlCell{Column: col, Value: val})
		}
		out.Rows = append(out.Rows, xr)
	}

	data, err := xml.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("xml format error: %w", err)
	}
	return xml.Header + string(data), nil
}

// YAMLFormatter formats results as YAML-like text.
type YAMLFormatter struct{}

// ContentType returns the MIME type for YAML.
func (f *YAMLFormatter) ContentType() string {
	return "application/yaml"
}

// Format converts the result to a YAML-like string.
func (f *YAMLFormatter) Format(result cypher.Result) (string, error) {
	var b strings.Builder

	if len(result.Columns) > 0 {
		b.WriteString("columns:\n")
		for _, col := range result.Columns {
			b.WriteString(fmt.Sprintf("  - %s\n", col))
		}
	}

	if len(result.Rows) > 0 {
		b.WriteString("rows:\n")
		for i, row := range result.Rows {
			b.WriteString(fmt.Sprintf("  - row: %d\n", i))
			for _, col := range result.Columns {
				val := "null"
				if v, ok := row[col]; ok {
					val = fmt.Sprintf("%v", v)
				}
				b.WriteString(fmt.Sprintf("    %s: %s\n", col, val))
			}
		}
	}

	if result.AffectedNodes > 0 || result.AffectedRels > 0 {
		b.WriteString("stats:\n")
		b.WriteString(fmt.Sprintf("  affected_nodes: %d\n", result.AffectedNodes))
		b.WriteString(fmt.Sprintf("  affected_rels: %d\n", result.AffectedRels))
		b.WriteString(fmt.Sprintf("  nodes_created: %d\n", result.NodesCreated))
		b.WriteString(fmt.Sprintf("  rels_created: %d\n", result.RelsCreated))
	}

	return b.String(), nil
}

// CSVFormatter formats results as CSV.
type CSVFormatter struct{}

// ContentType returns the MIME type for CSV.
func (f *CSVFormatter) ContentType() string {
	return "text/csv"
}

// Format converts the result to a CSV string.
func (f *CSVFormatter) Format(result cypher.Result) (string, error) {
	if len(result.Columns) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write header
	if err := w.Write(result.Columns); err != nil {
		return "", fmt.Errorf("csv header error: %w", err)
	}

	// Write rows
	for _, row := range result.Rows {
		record := make([]string, len(result.Columns))
		for i, col := range result.Columns {
			if v, ok := row[col]; ok {
				record[i] = fmt.Sprintf("%v", v)
			} else {
				record[i] = ""
			}
		}
		if err := w.Write(record); err != nil {
			return "", fmt.Errorf("csv row error: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("csv flush error: %w", err)
	}

	return buf.String(), nil
}

// TableFormatter formats results as a human-readable table (default).
type TableFormatter struct{}

// ContentType returns the MIME type for plain text.
func (f *TableFormatter) ContentType() string {
	return "text/plain"
}

// Format converts the result to a table string.
func (f *TableFormatter) Format(result cypher.Result) (string, error) {
	if len(result.Rows) == 0 {
		if result.AffectedNodes > 0 || result.AffectedRels > 0 {
			return fmt.Sprintf("Affected: %d nodes, %d relationships\n", result.AffectedNodes, result.AffectedRels), nil
		}
		return "(no results)\n", nil
	}

	var b strings.Builder

	// Calculate column widths
	widths := make(map[string]int)
	for _, col := range result.Columns {
		widths[col] = len(col)
	}
	for _, row := range result.Rows {
		for _, col := range result.Columns {
			val := ""
			if v, ok := row[col]; ok {
				val = fmt.Sprintf("%v", v)
			}
			if len(val) > widths[col] {
				widths[col] = len(val)
			}
		}
	}

	// Write header
	for _, col := range result.Columns {
		b.WriteString(fmt.Sprintf("%-*s  ", widths[col], col))
	}
	b.WriteString("\n")

	// Write separator
	for _, col := range result.Columns {
		b.WriteString(strings.Repeat("-", widths[col]))
		b.WriteString("  ")
	}
	b.WriteString("\n")

	// Write rows
	for _, row := range result.Rows {
		for _, col := range result.Columns {
			val := ""
			if v, ok := row[col]; ok {
				val = fmt.Sprintf("%v", v)
			}
			b.WriteString(fmt.Sprintf("%-*s  ", widths[col], val))
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}
