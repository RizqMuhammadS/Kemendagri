package services

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/meeting-minutes-ai/internal/config"
	"github.com/yourusername/meeting-minutes-ai/internal/dto"
)

// ExportService handles exporting meeting minutes to PDF and Word
type ExportService struct {
	cfg *config.Config
}

// NewExportService creates a new export service
func NewExportService(cfg *config.Config) *ExportService {
	return &ExportService{cfg: cfg}
}

// Export exports meeting minutes in the specified format
func (s *ExportService) Export(detail *dto.MeetingDetailResponse, format string) (string, error) {
	// Ensure export directory exists
	if err := os.MkdirAll(s.cfg.ExportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	switch strings.ToLower(format) {
	case "pdf":
		return s.exportPDF(detail)
	case "word", "docx":
		return s.exportWord(detail)
	default:
		return "", fmt.Errorf("unsupported export format: %s", format)
	}
}

func (s *ExportService) exportPDF(detail *dto.MeetingDetailResponse) (string, error) {
	// Generate HTML first, then we would convert to PDF
	htmlContent := s.generateHTML(detail)

	filename := fmt.Sprintf("notulensi_%s_%d.html", detail.Date, detail.ID)
	outputPath := filepath.Join(s.cfg.ExportDir, filename)

	if err := os.WriteFile(outputPath, []byte(htmlContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write export file: %w", err)
	}

	return outputPath, nil
}

func (s *ExportService) exportWord(detail *dto.MeetingDetailResponse) (string, error) {
	// Generate HTML that can be opened as Word document
	htmlContent := s.generateWordHTML(detail)

	filename := fmt.Sprintf("notulensi_%s_%d.doc", detail.Date, detail.ID)
	outputPath := filepath.Join(s.cfg.ExportDir, filename)

	if err := os.WriteFile(outputPath, []byte(htmlContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write export file: %w", err)
	}

	return outputPath, nil
}

func (s *ExportService) generateHTML(detail *dto.MeetingDetailResponse) string {
	tmpl := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>Notulensi Rapat - {{.Title}}</title>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; margin: 40px; color: #333; }
        h1 { color: #1a5276; border-bottom: 2px solid #2980b9; padding-bottom: 10px; }
        h2 { color: #2c3e50; margin-top: 30px; }
        .header { margin-bottom: 20px; }
        .header p { margin: 5px 0; }
        .summary { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .point { padding: 8px 0; border-bottom: 1px solid #eee; }
        .decision { padding: 8px 0; border-bottom: 1px solid #eee; }
        .action-item { padding: 10px; margin: 10px 0; background: #eaf2f8; border-left: 4px solid #2980b9; }
        .action-item .assignee { color: #7f8c8d; font-size: 0.9em; }
        .action-item .deadline { color: #e74c3c; font-size: 0.9em; }
        .participant { display: inline-block; margin: 5px; padding: 5px 10px; background: #d5f5e3; border-radius: 3px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #2980b9; color: white; }
        .footer { margin-top: 50px; text-align: center; color: #95a5a6; font-size: 0.8em; }
    </style>
</head>
<body>
    <h1>Notulensi Rapat</h1>
    <div class="header">
        <p><strong>Judul:</strong> {{.Title}}</p>
        <p><strong>Tanggal:</strong> {{.Date}}</p>
        <p><strong>Lokasi:</strong> {{.Location}}</p>
        <p><strong>Status:</strong> {{.Status}}</p>
    </div>

    <h2>Peserta Rapat</h2>
    <div>
        {{range .Participants}}
        <span class="participant">{{.Name}} ({{.Role}})</span>
        {{else}}
        <p>Tidak ada peserta tercatat</p>
        {{end}}
    </div>

    <h2>Ringkasan Rapat</h2>
    <div class="summary">
        <p>{{.Summary}}</p>
    </div>

    <h2>Poin Pembahasan</h2>
    <ol>
        {{range .DiscussionPoints}}
        <li class="point">{{.Point}}</li>
        {{else}}
        <p>Tidak ada poin pembahasan</p>
        {{end}}
    </ol>

    <h2>Keputusan</h2>
    <ul>
        {{range .Decisions}}
        <li class="decision">{{.Decision}}</li>
        {{else}}
        <p>Tidak ada keputusan</p>
        {{end}}
    </ul>

    <h2>Action Items</h2>
    {{range .ActionItems}}
    <div class="action-item">
        <p><strong>{{.Task}}</strong></p>
        <p class="assignee">Penanggung Jawab: {{.Assignee}}</p>
        <p class="deadline">Deadline: {{.Deadline}}</p>
    </div>
    {{else}}
    <p>Tidak ada action items</p>
    {{end}}

    <div class="footer">
        <p>Dokumen ini dibuat secara otomatis oleh Sistem Notulensi AI</p>
        <p>{{.CreatedAt}}</p>
    </div>
</body>
</html>`

	t, _ := template.New("minutes").Parse(tmpl)
	var buf bytes.Buffer
	_ = t.Execute(&buf, detail)
	return buf.String()
}

func (s *ExportService) generateWordHTML(detail *dto.MeetingDetailResponse) string {
	// Word-compatible HTML (uses mso styles for Word)
	tmpl := `<html xmlns:o='urn:schemas-microsoft-com:office:office'
      xmlns:w='urn:schemas-microsoft-com:office:word'
      xmlns='http://www.w3.org/TR/REC-html40'>
<head>
    <meta charset="UTF-8">
    <title>Notulensi Rapat - {{.Title}}</title>
    <!--[if gte mso 9]>
    <xml>
        <w:WordDocument>
            <w:View>Print</w:View>
        </w:WordDocument>
    </xml>
    <![endif]-->
    <style>
        body { font-family: 'Calibri', Arial, sans-serif; margin: 1in; }
        h1 { color: #1a5276; }
        h2 { color: #2c3e50; margin-top: 20pt; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8pt; }
        th { background: #2980b9; color: white; }
    </style>
</head>
<body>
    <h1>Notulensi Rapat</h1>
    <table>
        <tr><td><strong>Judul</strong></td><td>{{.Title}}</td></tr>
        <tr><td><strong>Tanggal</strong></td><td>{{.Date}}</td></tr>
        <tr><td><strong>Lokasi</strong></td><td>{{.Location}}</td></tr>
        <tr><td><strong>Status</strong></td><td>{{.Status}}</td></tr>
    </table>

    <h2>Peserta Rapat</h2>
    <table>
        <tr><th>Nama</th><th>Peran</th></tr>
        {{range .Participants}}
        <tr><td>{{.Name}}</td><td>{{.Role}}</td></tr>
        {{else}}
        <tr><td colspan="2">Tidak ada peserta</td></tr>
        {{end}}
    </table>

    <h2>Ringkasan Rapat</h2>
    <p>{{.Summary}}</p>

    <h2>Poin Pembahasan</h2>
    <ol>
        {{range .DiscussionPoints}}
        <li>{{.Point}}</li>
        {{else}}
        <p>Tidak ada poin pembahasan</p>
        {{end}}
    </ol>

    <h2>Keputusan</h2>
    <ul>
        {{range .Decisions}}
        <li>{{.Decision}}</li>
        {{else}}
        <p>Tidak ada keputusan</p>
        {{end}}
    </ul>

    <h2>Action Items</h2>
    <table>
        <tr><th>Tugas</th><th>Penanggung Jawab</th><th>Deadline</th></tr>
        {{range .ActionItems}}
        <tr><td>{{.Task}}</td><td>{{.Assignee}}</td><td>{{.Deadline}}</td></tr>
        {{else}}
        <tr><td colspan="3">Tidak ada action items</td></tr>
        {{end}}
    </table>

    <p><em>Dokumen dibuat: {{time.Now}}</em></p>
</body>
</html>`

	funcMap := template.FuncMap{
		"time": time.Now,
	}

	t, _ := template.New("word").Funcs(funcMap).Parse(tmpl)
	var buf bytes.Buffer
	_ = t.Execute(&buf, detail)
	return buf.String()
}