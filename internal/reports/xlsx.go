package reports

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

const maximumXLSXDataRows = 1_048_575

func WorkbookXLSX(ctx context.Context, source QuerySource, crawlID contracts.ID, output io.Writer) error {
	archive := zip.NewWriter(output)
	files := map[string]string{
		"[Content_Types].xml":        `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
		"_rels/.rels":                `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":            `<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Pages" sheetId="1" r:id="rId1"/><sheet name="Issues" sheetId="2" r:id="rId2"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/></Relationships>`,
	}
	for name, body := range files {
		if err := zipFile(archive, name, func(writer io.Writer) error { _, err := io.WriteString(writer, body); return err }); err != nil {
			return err
		}
	}
	if err := zipFile(archive, "xl/worksheets/sheet1.xml", func(writer io.Writer) error { return writePagesSheet(ctx, source, crawlID, writer) }); err != nil {
		return err
	}
	if err := zipFile(archive, "xl/worksheets/sheet2.xml", func(writer io.Writer) error { return writeIssuesSheet(ctx, source, crawlID, writer) }); err != nil {
		return err
	}
	return archive.Close()
}

func zipFile(archive *zip.Writer, name string, write func(io.Writer) error) error {
	writer, err := archive.Create(name)
	if err != nil {
		return err
	}
	return write(writer)
}

func writePagesSheet(ctx context.Context, source QuerySource, crawlID contracts.ID, output io.Writer) error {
	if _, err := io.WriteString(output, `<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`); err != nil {
		return err
	}
	row := 1
	if err := writeXMLRow(output, row, []string{"URL", "Status Code", "Depth", "Title", "Meta Description", "Canonical", "Robots", "Language", "Text Length", "Content Hash"}); err != nil {
		return err
	}
	row++
	cursor := ""
	for {
		page, err := source.ListPages(ctx, crawlID, contracts.PageRequest{Cursor: cursor, Limit: 1000})
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			if row > maximumXLSXDataRows+1 {
				return fmt.Errorf("XLSX row limit exceeded")
			}
			values := []string{item.URL, strconv.Itoa(item.StatusCode), strconv.Itoa(item.Depth), item.Title, item.MetaDescription, item.CanonicalURL, item.RobotsDirectives, item.Language, strconv.Itoa(item.TextLength), item.ContentHash}
			if err := writeXMLRow(output, row, values); err != nil {
				return err
			}
			row++
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	_, err := io.WriteString(output, `</sheetData></worksheet>`)
	return err
}
func writeIssuesSheet(ctx context.Context, source QuerySource, crawlID contracts.ID, output io.Writer) error {
	if _, err := io.WriteString(output, `<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`); err != nil {
		return err
	}
	row := 1
	if err := writeXMLRow(output, row, []string{"ID", "Rule ID", "Rule Version", "Subject Type", "Subject ID", "Severity", "Evidence", "Created At"}); err != nil {
		return err
	}
	row++
	cursor := ""
	for {
		page, err := source.ListIssues(ctx, crawlID, contracts.PageRequest{Cursor: cursor, Limit: 1000})
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			if row > maximumXLSXDataRows+1 {
				return fmt.Errorf("XLSX row limit exceeded")
			}
			values := []string{strconv.FormatInt(item.ID, 10), item.RuleID, strconv.Itoa(item.RuleVersion), item.SubjectType, item.SubjectID, item.Severity, item.EvidenceJSON, item.CreatedAt}
			if err := writeXMLRow(output, row, values); err != nil {
				return err
			}
			row++
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	_, err := io.WriteString(output, `</sheetData></worksheet>`)
	return err
}
func writeXMLRow(output io.Writer, row int, values []string) error {
	if _, err := fmt.Fprintf(output, `<row r="%d">`, row); err != nil {
		return err
	}
	for index, value := range values {
		column := xlsxColumn(index + 1)
		safe := spreadsheetSafe(value)
		var escaped bytes.Buffer
		if err := xml.EscapeText(&escaped, []byte(safe)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(output, `<c r="%s%d" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, column, row, escaped.String()); err != nil {
			return err
		}
	}
	_, err := io.WriteString(output, "</row>")
	return err
}
func xlsxColumn(value int) string {
	result := ""
	for value > 0 {
		value--
		result = string(rune('A'+value%26)) + result
		value /= 26
	}
	return result
}
