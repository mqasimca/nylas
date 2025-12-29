package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ResourceView interface for all views.
type ResourceView interface {
	Name() string
	Title() string
	Primitive() tview.Primitive
	Hints() []Hint
	Load()
	Refresh()
	Filter(string)
	HandleKey(*tcell.EventKey) *tcell.EventKey
}

// BaseTableView provides common table view functionality.
type BaseTableView struct {
	app    *App
	table  *Table
	name   string
	title  string
	hints  []Hint
	filter string
}

func newBaseTableView(app *App, name, title string) *BaseTableView {
	return &BaseTableView{
		app:   app,
		table: NewTable(app.styles),
		name:  name,
		title: title,
	}
}

func (v *BaseTableView) Name() string               { return v.name }
func (v *BaseTableView) Title() string              { return v.title }
func (v *BaseTableView) Primitive() tview.Primitive { return v.table }
func (v *BaseTableView) Hints() []Hint              { return v.hints }
func (v *BaseTableView) Filter(f string)            { v.filter = f }

func (v *BaseTableView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	return event // Let table handle navigation
}
