package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Language-specific regex patterns for top-level declarations.
var codePatterns = map[string][]*regexp.Regexp{
	".go": {
		regexp.MustCompile(`(?m)^func\s+(?:\([^)]*\)\s+)?(\w+)`),
		regexp.MustCompile(`(?m)^type\s+(\w+)`),
	},
	".py": {
		regexp.MustCompile(`(?m)^def\s+(\w+)`),
		regexp.MustCompile(`(?m)^class\s+(\w+)`),
	},
	".ts": {
		regexp.MustCompile(`(?m)^(?:export\s+)?(?:async\s+)?function\s+(\w+)`),
		regexp.MustCompile(`(?m)^(?:export\s+)?class\s+(\w+)`),
	},
	".js": {
		regexp.MustCompile(`(?m)^(?:export\s+)?(?:async\s+)?function\s+(\w+)`),
		regexp.MustCompile(`(?m)^(?:export\s+)?class\s+(\w+)`),
	},
	".tsx": {
		regexp.MustCompile(`(?m)^(?:export\s+)?(?:async\s+)?function\s+(\w+)`),
		regexp.MustCompile(`(?m)^(?:export\s+)?class\s+(\w+)`),
	},
	".jsx": {
		regexp.MustCompile(`(?m)^(?:export\s+)?(?:async\s+)?function\s+(\w+)`),
		regexp.MustCompile(`(?m)^(?:export\s+)?class\s+(\w+)`),
	},
	".rs": {
		regexp.MustCompile(`(?m)^(?:pub\s+)?fn\s+(\w+)`),
		regexp.MustCompile(`(?m)^(?:pub\s+)?struct\s+(\w+)`),
		regexp.MustCompile(`(?m)^impl\s+(\w+)`),
	},
}

func parseCode(filePath string, ext string) ([]Chunk, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	text := string(data)
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("file is empty: %s", filePath)
	}

	format := formatMap[ext]
	if format == "" {
		format = strings.TrimPrefix(ext, ".")
	}

	// If file is small, return as single chunk
	if len(text) < maxChunkSize {
		return []Chunk{{
			Title:      filepath.Base(filePath),
			Content:    text,
			Format:     format,
			ChunkIndex: 0,
		}}, nil
	}

	patterns, ok := codePatterns[ext]
	if !ok {
		// No patterns for this extension — return whole file
		return []Chunk{{
			Title:      filepath.Base(filePath),
			Content:    text,
			Format:     format,
			ChunkIndex: 0,
		}}, nil
	}

	// Find all declaration positions
	type decl struct {
		name string
		pos  int
	}
	var decls []decl

	for _, pat := range patterns {
		matches := pat.FindAllStringSubmatchIndex(text, -1)
		for _, m := range matches {
			name := ""
			// Find the first non-empty capture group
			for g := 2; g < len(m); g += 2 {
				if m[g] >= 0 {
					name = text[m[g]:m[g+1]]
					break
				}
			}
			if name != "" {
				decls = append(decls, decl{name: name, pos: m[0]})
			}
		}
	}

	if len(decls) == 0 {
		return []Chunk{{
			Title:      filepath.Base(filePath),
			Content:    text,
			Format:     format,
			ChunkIndex: 0,
		}}, nil
	}

	// Sort by position
	for i := 1; i < len(decls); i++ {
		for j := i; j > 0 && decls[j].pos < decls[j-1].pos; j-- {
			decls[j], decls[j-1] = decls[j-1], decls[j]
		}
	}

	// Extract chunks between declarations
	var chunks []Chunk
	for i, d := range decls {
		end := len(text)
		if i+1 < len(decls) {
			end = decls[i+1].pos
		}
		content := strings.TrimSpace(text[d.pos:end])
		if content == "" {
			continue
		}
		chunks = append(chunks, Chunk{
			Title:      d.name,
			Content:    content,
			Format:     format,
			ChunkIndex: len(chunks),
		})
	}

	// Prepend any content before first declaration (imports, package, etc.)
	if decls[0].pos > 0 {
		preamble := strings.TrimSpace(text[:decls[0].pos])
		if preamble != "" {
			pre := Chunk{
				Title:      "preamble",
				Content:    preamble,
				Format:     format,
				ChunkIndex: 0,
			}
			// Shift all chunk indices
			for i := range chunks {
				chunks[i].ChunkIndex = i + 1
			}
			chunks = append([]Chunk{pre}, chunks...)
		}
	}

	return chunks, nil
}
