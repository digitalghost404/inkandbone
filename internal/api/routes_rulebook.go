package api

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/digitalghost404/inkandbone/internal/db"
)

var errInvalidPDF = errors.New("invalid or unreadable PDF")

func (s *Server) handleListRulebookSources(w http.ResponseWriter, r *http.Request) {
	rulesetID, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid ruleset id", http.StatusBadRequest)
		return
	}
	sources, err := s.db.ListRulebookSources(rulesetID)
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if sources == nil {
		sources = []db.RulebookSource{}
	}
	writeJSON(w, sources)
}

func (s *Server) handleIngestRulebook(w http.ResponseWriter, r *http.Request) {
	rulesetID, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid ruleset id", http.StatusBadRequest)
		return
	}

	ct := r.Header.Get("Content-Type")

	var text string
	var source string

	switch {
	case strings.HasPrefix(ct, "text/plain"):
		source = r.URL.Query().Get("source")
		r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
			return
		}
		text = string(b)

	case strings.HasPrefix(ct, "multipart/form-data"):
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			http.Error(w, "parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		source = r.FormValue("source")
		file, _, err := r.FormFile("rulebook")
		if err != nil {
			http.Error(w, "rulebook field is required: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		extracted, err := extractPDFText(file)
		if err != nil {
			if errors.Is(err, errInvalidPDF) {
				http.Error(w, "pdf extraction: "+err.Error(), http.StatusUnprocessableEntity)
			} else {
				http.Error(w, "pdf extraction: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
		text = extracted

	default:
		http.Error(w, "unsupported Content-Type; use text/plain or multipart/form-data", http.StatusUnsupportedMediaType)
		return
	}

	if source == "" {
		source = "Core Rulebook"
	}

	chunks := chunkByHeadingsWithSource(text, source)
	// Replace only chunks for this source — other books are untouched.
	if err := s.db.DeleteRulebookChunksBySource(rulesetID, source); err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.db.CreateRulebookChunks(rulesetID, chunks); err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{"chunks_created": len(chunks), "source": source})
}

// chunkByHeadingsWithSource splits text into chunks delimited by lines starting with "#",
// tagging each chunk with the given source book name.
// If no heading lines are found, the entire text becomes one chunk with an empty heading.
func chunkByHeadingsWithSource(text, source string) []db.RulebookChunk {
	var chunks []db.RulebookChunk
	var currentHeading strings.Builder
	var currentContent strings.Builder

	flush := func() {
		if currentHeading.Len() > 0 || currentContent.Len() > 0 {
			chunks = append(chunks, db.RulebookChunk{
				Source:  source,
				Heading: strings.TrimSpace(strings.TrimLeft(currentHeading.String(), "#")),
				Content: strings.TrimSpace(currentContent.String()),
			})
			currentHeading.Reset()
			currentContent.Reset()
		}
	}

	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "#") {
			flush()
			currentHeading.WriteString(line)
		} else {
			currentContent.WriteString(line + "\n")
		}
	}
	flush()
	return chunks
}

// extractPDFText validates the PDF and extracts readable text from its content streams.
// It writes content streams to a temp directory, reads them, and extracts text
// from PDF text operators (Tj, TJ).
func extractPDFText(r io.ReadSeeker) (string, error) {
	conf := model.NewDefaultConfiguration()
	if err := pdfapi.Validate(r, conf); err != nil {
		return "", fmt.Errorf("%w: %v", errInvalidPDF, err)
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "rulebook-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	if err := pdfapi.ExtractContent(r, tmpDir, "rulebook.pdf", nil, conf); err != nil {
		return "", err
	}

	var sb strings.Builder
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		f, err := os.Open(filepath.Join(tmpDir, entry.Name()))
		if err != nil {
			return "", err
		}
		text := extractTextFromContentStream(f)
		f.Close()
		sb.WriteString(text)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

// extractTextFromContentStream reads a PDF content stream and extracts text
// from Tj and TJ text-showing operators.
func extractTextFromContentStream(r io.Reader) string {
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		// Extract text from Tj operator: (text) Tj
		if strings.HasSuffix(trimmed, "Tj") {
			if t := extractParenText(line); t != "" {
				sb.WriteString(t)
			}
		}
		// Extract text from TJ operator: [(text) ...] TJ
		if strings.HasSuffix(trimmed, "TJ") {
			if t := extractParenText(line); t != "" {
				sb.WriteString(t)
			}
		}
	}
	return sb.String()
}

// extractParenText extracts all text between unescaped parentheses in a line.
func extractParenText(line string) string {
	var sb strings.Builder
	inParen := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '\\' && i+1 < len(line) {
			if inParen {
				sb.WriteByte(line[i+1])
			}
			i++
			continue
		}
		if ch == '(' {
			inParen = true
			continue
		}
		if ch == ')' {
			inParen = false
			continue
		}
		if inParen {
			sb.WriteByte(ch)
		}
	}
	return sb.String()
}
