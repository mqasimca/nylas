package output

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mqasimca/nylas/internal/ports"
	"github.com/stretchr/testify/assert"
)

func TestJSONWriter_Write(t *testing.T) {
	type item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	err := jw.Write(item{ID: "123", Name: "Test"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"id": "123"`)
	assert.Contains(t, output, `"name": "Test"`)
}

func TestJSONWriter_WriteList(t *testing.T) {
	type item struct {
		ID string `json:"id"`
	}

	data := []item{{ID: "1"}, {ID: "2"}}
	columns := []ports.Column{{Header: "ID", Field: "ID"}}

	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	err := jw.WriteList(data, columns)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"id": "1"`)
	assert.Contains(t, output, `"id": "2"`)
}

func TestJSONWriter_WriteError(t *testing.T) {
	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	err := jw.WriteError(errors.New("test error"))
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"error": "test error"`)
}
