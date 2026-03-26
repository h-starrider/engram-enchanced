package parser

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempFile(t *testing.T, ext, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*"+ext)
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestParseTextFile(t *testing.T) {
	content := "First paragraph with enough content to be meaningful.\n\nSecond paragraph also has some text.\n\nThird paragraph completes the document."
	path := writeTempFile(t, ".txt", content)

	chunks, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse text: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}
	if chunks[0].Format != "txt" {
		t.Fatalf("expected format 'txt', got %q", chunks[0].Format)
	}
	// All paragraphs are short, should be in 1 chunk
	if !strings.Contains(chunks[0].Content, "First paragraph") {
		t.Fatalf("expected first paragraph in chunk, got %q", chunks[0].Content)
	}
}

func TestParseTextFileMultiChunk(t *testing.T) {
	// Create content that exceeds maxChunkSize (10000 chars)
	var parts []string
	para := strings.Repeat("This is a test sentence. ", 100) // ~2500 chars
	for i := 0; i < 10; i++ {
		parts = append(parts, para)
	}
	content := strings.Join(parts, "\n\n")
	path := writeTempFile(t, ".txt", content)

	chunks, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse text multi-chunk: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks for large text, got %d", len(chunks))
	}
	for _, c := range chunks {
		if c.Format != "txt" {
			t.Fatalf("expected format 'txt', got %q", c.Format)
		}
	}
}

func TestParseCodeFileGo(t *testing.T) {
	goCode := `package main

import "fmt"

func Hello() {
	fmt.Println("hello")
}

func World() {
	fmt.Println("world")
}

func Main() {
	Hello()
	World()
}
`
	// Make it large enough to trigger per-function chunking
	largeGoCode := goCode + strings.Repeat("\n// padding line to make file large\n", 500)
	path := writeTempFile(t, ".go", largeGoCode)

	chunks, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse Go: %v", err)
	}
	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks (3 functions), got %d", len(chunks))
	}

	// Check function names in titles
	titles := make(map[string]bool)
	for _, c := range chunks {
		titles[c.Title] = true
		if c.Format != "go" {
			t.Fatalf("expected format 'go', got %q", c.Format)
		}
	}
	for _, name := range []string{"Hello", "World", "Main"} {
		if !titles[name] {
			t.Errorf("expected chunk with title %q, titles: %v", name, titles)
		}
	}
}

func TestParseCodeFilePython(t *testing.T) {
	pyCode := strings.Repeat("# padding line for testing purposes\n", 500) + `
def hello():
    print("hello")
` + strings.Repeat("    # more padding\n", 200) + `
class MyClass:
    def method(self):
        pass

def world():
    print("world")
`
	path := writeTempFile(t, ".py", pyCode)

	chunks, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse Python: %v", err)
	}

	titles := make(map[string]bool)
	for _, c := range chunks {
		titles[c.Title] = true
	}
	for _, name := range []string{"hello", "MyClass"} {
		if !titles[name] {
			t.Errorf("expected chunk with title %q, titles: %v", name, titles)
		}
	}
}

func TestParseUnsupportedFormat(t *testing.T) {
	path := writeTempFile(t, ".bin", "binary data")

	_, err := Parse(path)
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := Parse("/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestFormatDetection(t *testing.T) {
	tests := []struct {
		ext    string
		format string
	}{
		{".pdf", "pdf"}, {".docx", "docx"}, {".xlsx", "xlsx"}, {".xls", "xlsx"},
		{".go", "go"}, {".py", "py"}, {".ts", "ts"}, {".js", "js"},
		{".tsx", "tsx"}, {".jsx", "jsx"}, {".rs", "rs"},
		{".txt", "txt"}, {".md", "md"}, {".log", "log"}, {".csv", "csv"},
		{".bin", ""}, {".exe", ""}, {"", ""},
	}
	for _, tt := range tests {
		got := FormatFor(tt.ext)
		if got != tt.format {
			t.Errorf("FormatFor(%q) = %q, want %q", tt.ext, got, tt.format)
		}
	}
}

func TestChunkIndexIsSequential(t *testing.T) {
	var parts []string
	para := strings.Repeat("Sequential test paragraph. ", 200) // ~5200 chars
	for i := 0; i < 5; i++ {
		parts = append(parts, para)
	}
	content := strings.Join(parts, "\n\n")
	path := writeTempFile(t, ".md", content)

	chunks, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if c.ChunkIndex != i {
			t.Errorf("chunk %d has ChunkIndex %d, expected %d", i, c.ChunkIndex, i)
		}
	}
}

func TestParseMarkdownFile(t *testing.T) {
	path := writeTempFile(t, ".md", "# Title\n\nSome markdown content here.\n\n## Section 2\n\nMore content.")
	chunks, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse markdown: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
	if chunks[0].Format != "md" {
		t.Fatalf("expected format 'md', got %q", chunks[0].Format)
	}
}

func TestParseSmallCodeFileReturnsSingleChunk(t *testing.T) {
	path := writeTempFile(t, ".go", "package main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n")
	chunks, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse small Go: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk for small file, got %d", len(chunks))
	}
	if chunks[0].Title != filepath.Base(path) {
		t.Fatalf("expected filename as title for small file, got %q", chunks[0].Title)
	}
}
