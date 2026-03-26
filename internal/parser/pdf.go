package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

func parsePDF(filePath string) ([]Chunk, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages == 0 {
		return nil, fmt.Errorf("pdf has no pages: %s", filePath)
	}

	var chunks []Chunk
	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		var buf bytes.Buffer
		texts := page.Content().Text
		for _, t := range texts {
			buf.WriteString(t.S)
		}

		content := strings.TrimSpace(buf.String())
		if content == "" {
			continue
		}

		chunks = append(chunks, Chunk{
			Title:      fmt.Sprintf("Page %d", i),
			Content:    content,
			Format:     "pdf",
			ChunkIndex: len(chunks),
		})
	}

	if len(chunks) == 0 {
		// Fallback: try GetPlainText for the whole document
		var buf bytes.Buffer
		plainReader, err := r.GetPlainText()
		if err == nil {
			buf.ReadFrom(plainReader)
			text := strings.TrimSpace(buf.String())
			if text != "" {
				chunks = append(chunks, Chunk{
					Title:      "Full Document",
					Content:    text,
					Format:     "pdf",
					ChunkIndex: 0,
				})
			}
		}
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no extractable text in pdf: %s", filePath)
	}

	return chunks, nil
}
