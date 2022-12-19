package pkg

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/atotto/clipboard"
	"github.com/olekukonko/tablewriter"
)

// Parser describes an interface to Parse an arbitrary document into
// our intermediate Content form.
type Parser interface {
	Parse(io.Reader) (Content, error)
}

// Content is the intermediate representation before it is converted
// to a table format.
type Content struct {
	header []string
	rows   [][]string
}

// Format converts the content of the reader to a table format using
// the supplied parser and writes it to the writer.
func Format(p Parser, r io.Reader, w io.Writer, enablePbcopy bool) error {
	c, err := p.Parse(r)
	if err != nil {
		return err
	}

	formatTable(c, w)

	if enablePbcopy {
		tsvPbcopy(c)
	}

	return nil
}

// tsv format to clipboard
func tsvPbcopy(c Content) {
	fmt.Println("\n📎 TSV RESULT")
	var tsv bytes.Buffer
	for _, head := range c.header {
		tsv.WriteString(head + "\t")
	}
	tsv.WriteString("\n")
	for _, row := range c.rows {
		for _, value := range row {
			tsv.WriteString(value + "\t")
		}
		tsv.WriteString("\n")
	}
	err := clipboard.WriteAll(tsv.String())
	if err != nil {
		panic(err)
	}
	fmt.Println("tsv format is saved into clipboard successfully.\nYou can now paste it into an excel sheet.")
}

// CSVParser is a parser implementation that parses CSV documents.
type CSVParser struct{}

// Parse converts the content of a reader to the Content representation.
func (c *CSVParser) Parse(reader io.Reader) (Content, error) {
	r := csv.NewReader(reader)

	header, err := r.Read()
	if err != nil {
		return Content{}, err
	}

	rows, err := r.ReadAll()
	if err != nil {
		return Content{}, err
	}

	return Content{
		header: header,
		rows:   rows,
	}, nil
}

// JSONParser is a parser implementation that parses JSON documents.
type JSONParser struct{}

// Parse converts the content of a reader to the Content representation.
func (j *JSONParser) Parse(reader io.Reader) (Content, error) {
	r := json.NewDecoder(reader)

	var rows []map[string]interface{}
	if err := r.Decode(&rows); err != nil {
		return Content{}, err
	}

	headers := collectHeader(rows)
	sort.Strings(headers)

	var outputRows [][]string
	for i, row := range rows {
		outputRow := make([]string, len(headers))
		for j, header := range headers {
			if j == 0 {
				outputRow[j] = strconv.Itoa(i + 1)
			} else {
				outputRow[j] = fmt.Sprintf("%v", row[header])
			}
		}
		outputRows = append(outputRows, outputRow)
	}

	return Content{
		header: headers,
		rows:   outputRows,
	}, nil
}

func formatTable(c Content, w io.Writer) {
	fmt.Printf("\n🕸️  TABLE RESULT (Rows:%d)\n", len(c.rows))
	table := tablewriter.NewWriter(w)
	table.SetHeader(c.header)
	table.AppendBulk(c.rows)
	table.Render()
}

func collectHeader(rows []map[string]interface{}) []string {
	headerMap := map[string]struct{}{}
	for _, row := range rows {
		for k := range row {
			headerMap[k] = struct{}{}
		}
	}

	var out []string
	out = append(out, "#")
	for header := range headerMap {
		out = append(out, header)
	}

	return out
}
