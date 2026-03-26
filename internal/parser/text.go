package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxChunkSize = 10000

func parseText(filePath string) ([]Chunk, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	text := string(data)
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("file is empty: %s", filePath)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	format := formatMap[ext]
	if format == "" {
		format = "txt"
	}

	paragraphs := strings.Split(text, "\n\n")

	var chunks []Chunk
	var current strings.Builder
	chunkIdx := 0

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if current.Len() > 0 && current.Len()+len(p)+2 > maxChunkSize {
			chunks = append(chunks, Chunk{
				Title:      fmt.Sprintf("Part %d", chunkIdx+1),
				Content:    current.String(),
				Format:     format,
				ChunkIndex: chunkIdx,
			})
			chunkIdx++
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(p)
	}

	if current.Len() > 0 {
		chunks = append(chunks, Chunk{
			Title:      fmt.Sprintf("Part %d", chunkIdx+1),
			Content:    current.String(),
			Format:     format,
			ChunkIndex: chunkIdx,
		})
	}

	return chunks, nil
}
