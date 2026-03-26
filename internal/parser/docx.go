package parser

import (
	"fmt"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

func parseDOCX(filePath string) ([]Chunk, error) {
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}
	defer r.Close()

	doc := r.Editable()
	content := doc.GetContent()

	// The content comes as XML; extract text between tags
	text := extractDocxText(content)
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("no extractable text in docx: %s", filePath)
	}

	// Chunk by paragraphs, merging up to maxChunkSize
	paragraphs := strings.Split(text, "\n")
	var chunks []Chunk
	var current strings.Builder
	chunkIdx := 0

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if current.Len() > 0 && current.Len()+len(p)+1 > maxChunkSize {
			chunks = append(chunks, Chunk{
				Title:      fmt.Sprintf("Section %d", chunkIdx+1),
				Content:    current.String(),
				Format:     "docx",
				ChunkIndex: chunkIdx,
			})
			chunkIdx++
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(p)
	}

	if current.Len() > 0 {
		chunks = append(chunks, Chunk{
			Title:      fmt.Sprintf("Section %d", chunkIdx+1),
			Content:    current.String(),
			Format:     "docx",
			ChunkIndex: chunkIdx,
		})
	}

	return chunks, nil
}

// extractDocxText strips XML tags and extracts text content.
func extractDocxText(xmlContent string) string {
	var result strings.Builder
	inTag := false
	lastWasBreak := false

	for i := 0; i < len(xmlContent); i++ {
		ch := xmlContent[i]
		if ch == '<' {
			// Check for paragraph/break tags
			remaining := xmlContent[i:]
			if strings.HasPrefix(remaining, "</w:p>") || strings.HasPrefix(remaining, "<w:br") {
				if !lastWasBreak {
					result.WriteString("\n")
					lastWasBreak = true
				}
			}
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteByte(ch)
			lastWasBreak = false
		}
	}
	return strings.TrimSpace(result.String())
}
