package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Column defines a table column.
type Column struct {
	Title  string
	Width  int  // Fixed width, 0 for auto
	Expand bool // Expand to fill space
}

// Table wraps tview.Table with k9s-style functionality.
type Table struct {
	*tview.Table
	styles        *Styles
	columns       []Column
	data          [][]string
	rowMeta       []RowMeta
	onSelect      func(*RowMeta)
	onDoubleClick func(*RowMeta)
}

// RowMeta holds metadata for a row.
type RowMeta struct {
	ID      string
	Data    any
	Unread  bool
	Starred bool
	Error   bool
}

// NewTable creates a new table component.
func NewTable(styles *Styles) *Table {
	t := &Table{
		Table:  tview.NewTable(),
		styles: styles,
	}

	t.SetBackgroundColor(styles.BgColor)
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetSelectable(true, false) // Row selectable, not cell
	t.SetFixed(1, 0)             // Fixed header row
	t.SetSeparator(' ')

	// Add border for k9s-style look
	t.SetBorder(true)
	t.SetBorderColor(styles.BorderColor)
	t.SetBorders(false) // Don't show internal cell borders

	// Selection style
	t.SetSelectedStyle(tcell.StyleDefault.
		Background(styles.TableSelectBg).
		Foreground(styles.TableSelectFg))

	// Set up selection changed callback
	t.SetSelectionChangedFunc(func(row, column int) {
		if t.onSelect != nil && row > 0 {
			if meta := t.SelectedMeta(); meta != nil {
				t.onSelect(meta)
			}
		}
	})

	return t
}

// SetOnSelect sets the callback for when a row is selected.
func (t *Table) SetOnSelect(handler func(*RowMeta)) {
	t.onSelect = handler
}

// SetOnDoubleClick sets the callback for double-click on a row.
func (t *Table) SetOnDoubleClick(handler func(*RowMeta)) {
	t.onDoubleClick = handler
}

// MouseHandler returns the mouse handler for the table.
func (t *Table) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return t.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Let the parent table handle most mouse events
		if handler := t.Table.MouseHandler(); handler != nil {
			consumed, capture = handler(action, event, setFocus)
		}

		// Handle double-click specifically
		if action == tview.MouseLeftDoubleClick {
			if t.onDoubleClick != nil {
				if meta := t.SelectedMeta(); meta != nil {
					t.onDoubleClick(meta)
				}
			}
			return true, nil
		}

		return consumed, capture
	})
}

// SetColumns sets the column definitions.
func (t *Table) SetColumns(cols []Column) {
	t.columns = cols
	t.renderHeader()
}

// SetData sets the table data.
func (t *Table) SetData(data [][]string, meta []RowMeta) {
	t.data = data
	t.rowMeta = meta
	t.render()
}

// SelectedMeta returns the metadata for the selected row.
func (t *Table) SelectedMeta() *RowMeta {
	row, _ := t.GetSelection()
	idx := row - 1 // Subtract header row
	if idx >= 0 && idx < len(t.rowMeta) {
		return &t.rowMeta[idx]
	}
	return nil
}

// GetSelectedRow returns the 0-based selected data row index.
func (t *Table) GetSelectedRow() int {
	row, _ := t.GetSelection()
	return row - 1 // Subtract header row
}

// GetRowCount returns the number of data rows.
func (t *Table) GetRowCount() int {
	return len(t.data)
}

func (t *Table) renderHeader() {
	headerColor := t.styles.TableHeaderFg

	for i, col := range t.columns {
		cell := tview.NewTableCell(col.Title).
			SetTextColor(headerColor).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold)

		if col.Width > 0 {
			cell.SetMaxWidth(col.Width)
		}
		if col.Expand {
			cell.SetExpansion(1)
		}

		t.SetCell(0, i, cell)
	}
}

func (t *Table) render() {
	// Clear existing data rows (keep header)
	rowCount := t.GetRowCount()
	for i := rowCount - 1; i >= 1; i-- {
		t.RemoveRow(i)
	}

	// Render data rows
	for rowIdx, row := range t.data {
		var meta RowMeta
		if rowIdx < len(t.rowMeta) {
			meta = t.rowMeta[rowIdx]
		}

		for colIdx, cell := range row {
			if colIdx >= len(t.columns) {
				continue
			}

			tableCell := tview.NewTableCell(cell).
				SetTextColor(t.styles.TableRowFg)

			// Status indicator in first column
			if colIdx == 0 && t.columns[0].Width <= 4 {
				tableCell = t.renderStatusCell(meta)
			}

			if t.columns[colIdx].Width > 0 {
				tableCell.SetMaxWidth(t.columns[colIdx].Width)
			}
			if t.columns[colIdx].Expand {
				tableCell.SetExpansion(1)
			}

			t.SetCell(rowIdx+1, colIdx, tableCell)
		}
	}

	// Select first data row if exists
	if len(t.data) > 0 {
		t.Select(1, 0)
	}
}

func (t *Table) renderStatusCell(meta RowMeta) *tview.TableCell {
	var indicator string
	var color tcell.Color

	if meta.Error {
		indicator = "✗"
		color = t.styles.ErrorColor
	} else if meta.Unread && meta.Starred {
		indicator = "●★"
		color = t.styles.InfoColor
	} else if meta.Unread {
		indicator = "●"
		color = t.styles.InfoColor
	} else if meta.Starred {
		indicator = "★"
		color = t.styles.WarnColor
	} else {
		indicator = " "
		color = t.styles.TableRowFg
	}

	return tview.NewTableCell(indicator).
		SetTextColor(color).
		SetAlign(tview.AlignCenter)
}
