// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

// Package xlsx writes Excel workbooks a row at a time, so large sheets never sit in memory.
package xlsx

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/autobrr/autobrr/pkg/errors"
)

// MaxLinks is how many hyperlinks Excel accepts in one worksheet; later link cells are written as plain text.
const MaxLinks = 65530

// Column is a sheet column with its header and width in characters.
type Column struct {
	Name  string
	Width float64
}

type cellKind int

const (
	kindEmpty cellKind = iota
	kindString
	kindNumber
	kindDecimal
	kindDate
	kindLink
)

// Cell is one value in a row.
type Cell struct {
	kind   cellKind
	text   string
	number float64
	date   time.Time
}

// Empty is a blank cell.
var Empty = Cell{}

// String is a text cell; an empty string is a blank cell.
func String(s string) Cell {
	if s == "" {
		return Empty
	}
	return Cell{kind: kindString, text: s}
}

// Int is a whole number cell; zero is a blank cell, since zero means unset in release fields.
func Int(n int) Cell {
	if n == 0 {
		return Empty
	}
	return Cell{kind: kindNumber, number: float64(n)}
}

// Decimal is a number cell shown with one decimal and thousands separators.
func Decimal(f float64) Cell {
	return Cell{kind: kindDecimal, number: f}
}

// Date is a date and time cell; Excel has no time zones, so it shows t's wall clock time.
func Date(t time.Time) Cell {
	if t.IsZero() {
		return Empty
	}
	return Cell{kind: kindDate, date: t}
}

// Link is a clickable URL cell.
func Link(url string) Cell {
	if url == "" {
		return Empty
	}
	return Cell{kind: kindLink, text: url}
}

// Style indexes into the cellXfs list in styles.xml.
const (
	styleHeader  = 1
	styleDate    = 2
	styleDecimal = 3
	styleLink    = 4
)

type link struct {
	ref string
	url string
}

type sheet struct {
	name    string
	columns int
	rows    int
}

// Writer writes a workbook to an io.Writer. Sheets are written one after another: AddSheet closes the previous one.
type Writer struct {
	zip    *zip.Writer
	out    *bufio.Writer
	sheets []sheet
	links  []link
	err    error
}

// NewWriter returns a Writer that writes the workbook to w. Close must be called to finish it.
func NewWriter(w io.Writer) *Writer {
	return &Writer{zip: zip.NewWriter(w)}
}

// AddSheet starts a new sheet with a bold, frozen header row and filter buttons on every column.
func (w *Writer) AddSheet(name string, columns []Column) error {
	if w.err != nil {
		return w.err
	}
	if err := w.closeSheet(); err != nil {
		return err
	}

	f, err := w.zip.Create(fmt.Sprintf("xl/worksheets/sheet%d.xml", len(w.sheets)+1))
	if err != nil {
		return w.fail(err)
	}

	w.out = bufio.NewWriter(f)
	w.sheets = append(w.sheets, sheet{name: name, columns: len(columns)})
	w.links = nil

	w.write(xml.Header)
	w.write(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	w.write(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
	w.write(`<cols>`)
	for i, column := range columns {
		w.write(fmt.Sprintf(`<col min="%d" max="%d" width="%s" customWidth="1"/>`, i+1, i+1, strconv.FormatFloat(column.Width, 'f', -1, 64)))
	}
	w.write(`</cols><sheetData>`)

	header := make([]Cell, len(columns))
	for i, column := range columns {
		header[i] = String(column.Name)
	}
	w.row(header, styleHeader)

	return w.err
}

// WriteRow appends a row to the current sheet.
func (w *Writer) WriteRow(cells ...Cell) error {
	if w.err == nil && w.out == nil {
		w.err = errors.New("xlsx: no sheet added")
	}
	if w.err != nil {
		return w.err
	}

	w.row(cells, 0)

	return w.err
}

// Close finishes the current sheet and the workbook. It does not close the underlying writer.
func (w *Writer) Close() error {
	if w.err != nil {
		return w.err
	}
	if len(w.sheets) == 0 {
		return errors.New("xlsx: workbook has no sheets")
	}
	if err := w.closeSheet(); err != nil {
		return err
	}

	files := []struct {
		name    string
		content string
	}{
		{"[Content_Types].xml", w.contentTypes()},
		{"_rels/.rels", rootRels},
		{"xl/workbook.xml", w.workbook()},
		{"xl/_rels/workbook.xml.rels", w.workbookRels()},
		{"xl/styles.xml", styles},
	}

	for _, file := range files {
		f, err := w.zip.Create(file.name)
		if err != nil {
			return w.fail(err)
		}
		if _, err := io.WriteString(f, file.content); err != nil {
			return w.fail(err)
		}
	}

	return w.fail(w.zip.Close())
}

func (w *Writer) row(cells []Cell, style int) {
	current := &w.sheets[len(w.sheets)-1]
	current.rows++

	w.write(fmt.Sprintf(`<row r="%d">`, current.rows))

	for i, cell := range cells {
		ref := ColumnName(i) + strconv.Itoa(current.rows)

		switch cell.kind {
		case kindEmpty:
			continue
		case kindString:
			w.inline(ref, cell.text, style)
		case kindNumber:
			w.write(fmt.Sprintf(`<c r="%s"><v>%s</v></c>`, ref, strconv.FormatFloat(cell.number, 'f', -1, 64)))
		case kindDecimal:
			w.write(fmt.Sprintf(`<c r="%s" s="%d"><v>%s</v></c>`, ref, styleDecimal, strconv.FormatFloat(cell.number, 'f', -1, 64)))
		case kindDate:
			w.write(fmt.Sprintf(`<c r="%s" s="%d"><v>%s</v></c>`, ref, styleDate, strconv.FormatFloat(serial(cell.date), 'f', -1, 64)))
		case kindLink:
			if len(w.links) < MaxLinks {
				w.links = append(w.links, link{ref: ref, url: cell.text})
				w.inline(ref, cell.text, styleLink)
			} else {
				w.inline(ref, cell.text, 0)
			}
		}
	}

	w.write(`</row>`)
}

func (w *Writer) inline(ref, text string, style int) {
	if style != 0 {
		w.write(fmt.Sprintf(`<c r="%s" s="%d" t="inlineStr"><is><t xml:space="preserve">`, ref, style))
	} else {
		w.write(fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t xml:space="preserve">`, ref))
	}
	w.write(escape(text))
	w.write(`</t></is></c>`)
}

func (w *Writer) closeSheet() error {
	if w.out == nil {
		return w.err
	}

	current := w.sheets[len(w.sheets)-1]
	index := len(w.sheets)

	w.write(`</sheetData>`)
	w.write(fmt.Sprintf(`<autoFilter ref="%s"/>`, filterRange(current)))

	if len(w.links) > 0 {
		w.write(`<hyperlinks>`)
		for i, l := range w.links {
			w.write(fmt.Sprintf(`<hyperlink ref="%s" r:id="rId%d"/>`, l.ref, i+1))
		}
		w.write(`</hyperlinks>`)
	}

	w.write(`</worksheet>`)

	if w.err == nil {
		w.err = w.out.Flush()
	}
	w.out = nil

	if w.err != nil || len(w.links) == 0 {
		return w.err
	}

	f, err := w.zip.Create(fmt.Sprintf("xl/worksheets/_rels/sheet%d.xml.rels", index))
	if err != nil {
		return w.fail(err)
	}

	var b strings.Builder
	b.WriteString(xml.Header)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i, l := range w.links {
		fmt.Fprintf(&b, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="%s" TargetMode="External"/>`, i+1, escape(l.url))
	}
	b.WriteString(`</Relationships>`)

	if _, err := io.WriteString(f, b.String()); err != nil {
		return w.fail(err)
	}

	w.links = nil

	return nil
}

func (w *Writer) write(s string) {
	if w.err == nil {
		_, w.err = w.out.WriteString(s)
	}
}

func (w *Writer) fail(err error) error {
	if err != nil && w.err == nil {
		w.err = errors.Wrap(err, "xlsx")
	}
	return w.err
}

func (w *Writer) contentTypes() string {
	var b strings.Builder
	b.WriteString(xml.Header)
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	b.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	b.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	for i := range w.sheets {
		fmt.Fprintf(&b, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i+1)
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func (w *Writer) workbook() string {
	var b strings.Builder
	b.WriteString(xml.Header)
	b.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for i, s := range w.sheets {
		fmt.Fprintf(&b, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, escape(s.name), i+1, i+1)
	}
	b.WriteString(`</sheets><definedNames>`)
	// Excel expects each autoFilter to have this hidden name and repairs the file without it.
	for i, s := range w.sheets {
		fmt.Fprintf(&b, `<definedName name="_xlnm._FilterDatabase" localSheetId="%d" hidden="1">'%s'!%s</definedName>`, i, escape(strings.ReplaceAll(s.name, "'", "''")), absolute(filterRange(s)))
	}
	b.WriteString(`</definedNames></workbook>`)
	return b.String()
}

func (w *Writer) workbookRels() string {
	var b strings.Builder
	b.WriteString(xml.Header)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := range w.sheets {
		fmt.Fprintf(&b, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i+1, i+1)
	}
	fmt.Fprintf(&b, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, len(w.sheets)+1)
	b.WriteString(`</Relationships>`)
	return b.String()
}

// ColumnName returns the letters Excel uses for the zero-based column index: A, B, ... Z, AA.
func ColumnName(index int) string {
	name := ""
	for index >= 0 {
		name = string(rune('A'+index%26)) + name
		index = index/26 - 1
	}
	return name
}

func filterRange(s sheet) string {
	return fmt.Sprintf("A1:%s%d", ColumnName(max(s.columns, 1)-1), s.rows)
}

func absolute(ref string) string {
	from, to, _ := strings.Cut(ref, ":")
	return dollar(from) + ":" + dollar(to)
}

func dollar(ref string) string {
	i := strings.IndexAny(ref, "0123456789")
	return "$" + ref[:i] + "$" + ref[i:]
}

// serial converts t's wall clock time to an Excel date: days since 1899-12-30.
func serial(t time.Time) float64 {
	wall := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
	return wall.Sub(excelEpoch).Seconds() / 86400
}

var excelEpoch = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)

func escape(s string) string {
	var b strings.Builder
	// EscapeText replaces characters XML cannot hold, like most control characters, so a stray byte in a release
	// name cannot corrupt the sheet.
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

const rootRels = xml.Header + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`

const styles = xml.Header + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
	`<numFmts count="2"><numFmt numFmtId="164" formatCode="yyyy-mm-dd hh:mm:ss"/><numFmt numFmtId="165" formatCode="#,##0.0"/></numFmts>` +
	`<fonts count="3"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font><font><u/><sz val="11"/><color rgb="FF0563C1"/><name val="Calibri"/></font></fonts>` +
	`<fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill></fills>` +
	`<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>` +
	`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
	`<cellXfs count="5">` +
	`<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
	`<xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0" applyFont="1"/>` +
	`<xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>` +
	`<xf numFmtId="165" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>` +
	`<xf numFmtId="0" fontId="2" fillId="0" borderId="0" xfId="0" applyFont="1"/>` +
	`</cellXfs>` +
	`<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>` +
	`</styleSheet>`
