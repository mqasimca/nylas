//go:build !integration

package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripHTMLForTUI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text unchanged",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "basic HTML tags removed",
			input:    "<p>Hello</p><p>World</p>",
			expected: "Hello\n\nWorld",
		},
		{
			name:     "br tags become newlines",
			input:    "Line 1<br>Line 2<br/>Line 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "div tags become newlines",
			input:    "<div>Section 1</div><div>Section 2</div>",
			expected: "Section 1\n\nSection 2",
		},
		{
			name:     "inline tags removed",
			input:    "<b>Bold</b> and <i>italic</i>",
			expected: "Bold and italic",
		},
		{
			name:     "anchor tags removed",
			input:    "Visit <a href=\"http://example.com\">example</a>",
			expected: "Visit example",
		},
		{
			name:     "style tags and content removed",
			input:    "<style>body { color: red; }</style>Hello",
			expected: "Hello",
		},
		{
			name:     "script tags and content removed",
			input:    "<script>alert('xss')</script>Safe content",
			expected: "Safe content",
		},
		{
			name:     "head tags and content removed",
			input:    "<head><title>Page</title></head><body>Content</body>",
			expected: "Content",
		},
		{
			name:     "HTML entities decoded",
			input:    "&lt;tag&gt; &amp; &quot;quoted&quot;",
			expected: "<tag> & \"quoted\"",
		},
		{
			name:     "multiple spaces collapsed",
			input:    "Hello    World",
			expected: "Hello World",
		},
		{
			name:     "multiple newlines collapsed",
			input:    "Line 1\n\n\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "windows newlines normalized",
			input:    "Line 1\r\nLine 2\rLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "whitespace trimmed from lines",
			input:    "  Hello  \n  World  ",
			expected: "Hello\nWorld",
		},
		{
			name:     "uppercase HTML tags handled",
			input:    "<P>Paragraph</P><BR>Line",
			expected: "Paragraph\n\nLine",
		},
		{
			name:     "list items become newlines",
			input:    "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expected: "Item 1\n\nItem 2",
		},
		{
			name:     "table rows become newlines",
			input:    "<table><tr><td>Cell 1</td></tr><tr><td>Cell 2</td></tr></table>",
			expected: "Cell 1\n\nCell 2",
		},
		{
			name:     "headers become newlines",
			input:    "<h1>Title</h1><h2>Subtitle</h2>",
			expected: "Title\n\nSubtitle",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only tags",
			input:    "<div><span></span></div>",
			expected: "",
		},
		{
			name:     "nested tags",
			input:    "<div><p><span>Deep</span></p></div>",
			expected: "Deep",
		},
		{
			name:     "self-closing br variants",
			input:    "A<br>B<br/>C<br />D",
			expected: "A\nB\nC\nD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTMLForTUI(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveTagWithContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		tag      string
		expected string
	}{
		{
			name:     "remove style tag",
			input:    "<style>body { color: red; }</style>Content",
			tag:      "style",
			expected: "Content",
		},
		{
			name:     "remove script tag",
			input:    "Before<script>alert('test')</script>After",
			tag:      "script",
			expected: "BeforeAfter",
		},
		{
			name:     "remove head tag",
			input:    "<head><title>Page</title><meta/></head>Body",
			tag:      "head",
			expected: "Body",
		},
		{
			name:     "remove multiple instances",
			input:    "<style>a</style>Mid<style>b</style>End",
			tag:      "style",
			expected: "MidEnd",
		},
		{
			name:     "case insensitive",
			input:    "<STYLE>CSS</STYLE>Content",
			tag:      "style",
			expected: "Content",
		},
		{
			name:     "tag not found",
			input:    "No tags here",
			tag:      "style",
			expected: "No tags here",
		},
		{
			name:     "empty input",
			input:    "",
			tag:      "style",
			expected: "",
		},
		{
			name:     "unclosed tag removed",
			input:    "<style>unclosed content<p>Keep this</p>",
			tag:      "style",
			expected: "unclosed content<p>Keep this</p>",
		},
		{
			name:     "tag with attributes",
			input:    "<style type=\"text/css\">css</style>Content",
			tag:      "style",
			expected: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeTagWithContent(tt.input, tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}
