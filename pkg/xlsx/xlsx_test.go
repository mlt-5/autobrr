// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package xlsx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readWorkbook(t *testing.T, data []byte) map[string]string {
	t.Helper()

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	files := map[string]string{}
	for _, f := range r.File {
		rc, err := f.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())

		decoder := xml.NewDecoder(bytes.NewReader(content))
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			require.NoError(t, err, "%s is not well-formed XML", f.Name)
		}

		files[f.Name] = string(content)
	}

	return files
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	require.NoError(t, w.AddSheet("Releases", []Column{{Name: "Time", Width: 19}, {Name: "Name", Width: 40}, {Name: "Size", Width: 10}, {Name: "Season", Width: 8}, {Name: "URL", Width: 30}}))
	require.NoError(t, w.WriteRow(
		Date(time.Date(2026, 10, 23, 21, 14, 7, 0, time.UTC)),
		String("Tom & Jerry <S01> \"quoted\" 🎬\x01"),
		Decimal(78431.2),
		Int(0),
		Link("https://example.com/t?id=1&x=2"),
	))
	require.NoError(t, w.AddSheet("Actions", []Column{{Name: "ID", Width: 8}}))
	require.NoError(t, w.WriteRow(Int(1842)))
	require.NoError(t, w.Close())

	files := readWorkbook(t, buf.Bytes())

	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/styles.xml", "xl/worksheets/sheet1.xml", "xl/worksheets/sheet2.xml", "xl/worksheets/_rels/sheet1.xml.rels"} {
		assert.Contains(t, files, name)
	}
	assert.NotContains(t, files, "xl/worksheets/_rels/sheet2.xml.rels")

	sheet := files["xl/worksheets/sheet1.xml"]
	assert.Contains(t, sheet, `<c r="A1" s="1" t="inlineStr"><is><t xml:space="preserve">Time</t></is></c>`)
	assert.Contains(t, sheet, `<c r="A2" s="2"><v>46318.88480324074</v></c>`)
	assert.Contains(t, sheet, `Tom &amp; Jerry &lt;S01&gt; &#34;quoted&#34; 🎬`)
	assert.Contains(t, sheet, `<c r="C2" s="3"><v>78431.2</v></c>`)
	assert.NotContains(t, sheet, `r="D2"`, "zero ints are blank")
	assert.Contains(t, sheet, `<c r="E2" s="4" t="inlineStr"><is><t xml:space="preserve">https://example.com/t?id=1&amp;x=2</t></is></c>`)
	assert.Contains(t, sheet, `<autoFilter ref="A1:E2"/>`)
	assert.Contains(t, sheet, `<hyperlink ref="E2" r:id="rId1"/>`)
	assert.Contains(t, sheet, `state="frozen"`)

	assert.Contains(t, files["xl/worksheets/_rels/sheet1.xml.rels"], `Target="https://example.com/t?id=1&amp;x=2" TargetMode="External"`)
	assert.Contains(t, files["xl/worksheets/sheet2.xml"], `<c r="A2"><v>1842</v></c>`)
	assert.Contains(t, files["xl/workbook.xml"], `<definedName name="_xlnm._FilterDatabase" localSheetId="1" hidden="1">'Actions'!$A$1:$A$2</definedName>`)
}

func TestWriterLinkLimit(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	require.NoError(t, w.AddSheet("Links", []Column{{Name: "URL", Width: 30}}))
	for i := 0; i < MaxLinks+2; i++ {
		require.NoError(t, w.WriteRow(Link("https://example.com")))
	}
	require.NoError(t, w.Close())

	files := readWorkbook(t, buf.Bytes())
	assert.Equal(t, MaxLinks, strings.Count(files["xl/worksheets/sheet1.xml"], "<hyperlink "))
}

func TestWriterErrors(t *testing.T) {
	w := NewWriter(io.Discard)
	assert.Error(t, w.WriteRow(String("x")), "rows need a sheet")

	assert.Error(t, NewWriter(io.Discard).Close(), "a workbook needs a sheet")
}

func TestColumnName(t *testing.T) {
	for index, want := range map[int]string{0: "A", 25: "Z", 26: "AA", 51: "AZ", 52: "BA", 701: "ZZ", 702: "AAA"} {
		assert.Equal(t, want, ColumnName(index))
	}
}
