// Package parser extracts text content from files in various formats.
// All parsers are pure Go with no CGO dependencies.
package parser

import (
	"errors"
	"path/filepath"
	"strings"
)

// Chunk represents a piece of extracted content from a file.
type Chunk struct {
	Title      string
	Content    string
	Format     string
	ChunkIndex int
}

// ErrUnsupportedFormat is returned when the file extension is not recognized.
var ErrUnsupportedFormat = errors.New("unsupported file format")

// formatMap maps file extensions to format categories.
var formatMap = map[string]string{
	".pdf":  "pdf",
	".docx": "docx",
	".xlsx": "xlsx",
	".xls":  "xlsx",
	".go":   "go",
	".py":   "py",
	".ts":   "ts",
	".js":   "js",
	".tsx":  "tsx",
	".jsx":  "jsx",
	".rs":   "rs",
	".txt":  "txt",
	".md":   "md",
	".log":  "log",
	".csv":  "csv",
}

// codeExtensions are extensions handled by the code parser.
var codeExtensions = map[string]bool{
	".go": true, ".py": true, ".ts": true, ".js": true,
	".tsx": true, ".jsx": true, ".rs": true,
}

// textExtensions are extensions handled by the text parser.
var textExtensions = map[string]bool{
	".txt": true, ".md": true, ".log": true, ".csv": true,
}

// FormatFor returns the format string for the given file extension, or "" if unsupported.
func FormatFor(ext string) string {
	return formatMap[strings.ToLower(ext)]
}

// Parse detects the file format by extension and extracts text chunks.
func Parse(filePath string) ([]Chunk, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" || formatMap[ext] == "" {
		return nil, ErrUnsupportedFormat
	}

	switch {
	case ext == ".pdf":
		return parsePDF(filePath)
	case ext == ".docx":
		return parseDOCX(filePath)
	case ext == ".xlsx" || ext == ".xls":
		return parseXLSX(filePath)
	case codeExtensions[ext]:
		return parseCode(filePath, ext)
	case textExtensions[ext]:
		return parseText(filePath)
	default:
		return nil, ErrUnsupportedFormat
	}
}
