package render

import (
	"encoding/json"
	"io"

	"github.com/sysmon/system-monitor-cli/internal/models"
)

// JSONRenderer renders metrics as JSON output
type JSONRenderer struct {
	writer  io.Writer
	encoder *json.Encoder
}

// NewJSONRenderer creates a new JSON renderer
func NewJSONRenderer(writer io.Writer) *JSONRenderer {
	encoder := json.NewEncoder(writer)
	return &JSONRenderer{
		writer:  writer,
		encoder: encoder,
	}
}

// Render serializes metrics to JSON and writes one line
func (r *JSONRenderer) Render(metrics *models.Metrics) error {
	return r.encoder.Encode(metrics)
}

// Clear is a no-op for JSON renderer
func (r *JSONRenderer) Clear() error {
	return nil
}

// Close is a no-op for JSON renderer
func (r *JSONRenderer) Close() error {
	return nil
}
