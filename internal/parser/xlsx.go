package parser

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

func parseXLSX(filePath string) ([]Chunk, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("xlsx has no sheets: %s", filePath)
	}

	var chunks []Chunk
	for i, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		if len(rows) == 0 {
			continue
		}

		var content strings.Builder
		for _, row := range rows {
			content.WriteString(strings.Join(row, " | "))
			content.WriteString("\n")
		}

		text := strings.TrimSpace(content.String())
		if text == "" {
			continue
		}

		chunks = append(chunks, Chunk{
			Title:      fmt.Sprintf("Sheet: %s", sheet),
			Content:    text,
			Format:     "xlsx",
			ChunkIndex: i,
		})
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no data in xlsx sheets: %s", filePath)
	}

	return chunks, nil
}
