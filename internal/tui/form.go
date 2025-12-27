package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FieldType defines the type of form field.
type FieldType int

const (
	FieldText FieldType = iota
	FieldTextArea
	FieldDropdown
	FieldCheckbox
	FieldDate
	FieldTime
	FieldDateTime
)

// FormField defines a field in a form.
type FormField struct {
	Key         string             // Unique key for the field
	Label       string             // Display label
	Type        FieldType          // Field type
	Placeholder string             // Placeholder text
	Value       string             // Initial value
	Options     []string           // Options for dropdown
	Required    bool               // Whether field is required
	Validator   func(string) error // Optional validation function
}

// Form provides a modal form for creating/editing resources.
type Form struct {
	*tview.Flex
	app      *App
	form     *tview.Form
	title    string
	fields   []FormField
	values   map[string]string
	onSubmit func(values map[string]string)
	onCancel func()
}

// NewForm creates a new form with the given fields.
func NewForm(app *App, title string, fields []FormField, onSubmit func(map[string]string), onCancel func()) *Form {
	f := &Form{
		Flex:     tview.NewFlex(),
		app:      app,
		title:    title,
		fields:   fields,
		values:   make(map[string]string),
		onSubmit: onSubmit,
		onCancel: onCancel,
	}

	f.init()
	return f
}

func (f *Form) init() {
	styles := f.app.styles

	f.form = tview.NewForm()
	f.form.SetBackgroundColor(styles.BgColor)
	f.form.SetFieldBackgroundColor(styles.BgColor)
	f.form.SetFieldTextColor(styles.FgColor)
	f.form.SetLabelColor(styles.TitleFg)
	f.form.SetButtonBackgroundColor(styles.FocusColor)
	f.form.SetButtonTextColor(styles.BgColor)
	f.form.SetBorder(true)
	f.form.SetBorderColor(styles.FocusColor)
	f.form.SetTitle(fmt.Sprintf(" %s ", f.title))
	f.form.SetTitleColor(styles.TitleFg)

	// Add fields
	for _, field := range f.fields {
		f.addField(field)
	}

	// Add buttons
	f.form.AddButton("Submit", f.submit)
	f.form.AddButton("Cancel", f.cancel)

	// Set up key capture
	f.form.SetInputCapture(f.handleInput)

	// Center the form
	f.SetDirection(tview.FlexRow)
	f.AddItem(nil, 0, 1, false)
	f.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(f.form, 60, 0, true).
		AddItem(nil, 0, 1, false), 0, 3, true)
	f.AddItem(nil, 0, 1, false)
}

func (f *Form) addField(field FormField) {
	// Set initial value
	f.values[field.Key] = field.Value

	switch field.Type {
	case FieldText:
		f.form.AddInputField(field.Label, field.Value, 40, nil, func(text string) {
			f.values[field.Key] = text
		})

	case FieldTextArea:
		f.form.AddTextArea(field.Label, field.Value, 40, 4, 0, func(text string) {
			f.values[field.Key] = text
		})

	case FieldDropdown:
		initialIndex := 0
		for i, opt := range field.Options {
			if opt == field.Value {
				initialIndex = i
				break
			}
		}
		f.form.AddDropDown(field.Label, field.Options, initialIndex, func(option string, _ int) {
			f.values[field.Key] = option
		})

	case FieldCheckbox:
		checked := field.Value == "true" || field.Value == "1"
		f.form.AddCheckbox(field.Label, checked, func(checked bool) {
			if checked {
				f.values[field.Key] = "true"
			} else {
				f.values[field.Key] = "false"
			}
		})

	case FieldDate, FieldTime, FieldDateTime:
		// Use text input with placeholder hint in label
		hint := field.Placeholder
		if hint == "" {
			switch field.Type {
			case FieldDate:
				hint = "YYYY-MM-DD"
			case FieldTime:
				hint = "HH:MM"
			case FieldDateTime:
				hint = "YYYY-MM-DD HH:MM"
			}
		}
		labelWithHint := fmt.Sprintf("%s (%s)", field.Label, hint)
		f.form.AddInputField(labelWithHint, field.Value, 20, nil, func(text string) {
			f.values[field.Key] = text
		})
	}
}

func (f *Form) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		f.cancel()
		return nil
	case tcell.KeyCtrlS:
		f.submit()
		return nil
	}
	return event
}

func (f *Form) submit() {
	// Validate required fields
	var errors []string
	for _, field := range f.fields {
		value := strings.TrimSpace(f.values[field.Key])
		if field.Required && value == "" {
			errors = append(errors, fmt.Sprintf("%s is required", field.Label))
		}
		if field.Validator != nil && value != "" {
			if err := field.Validator(value); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %s", field.Label, err.Error()))
			}
		}
	}

	if len(errors) > 0 {
		f.app.Flash(FlashError, "%s", strings.Join(errors, "; "))
		return
	}

	if f.onSubmit != nil {
		f.onSubmit(f.values)
	}
}

func (f *Form) cancel() {
	if f.onCancel != nil {
		f.onCancel()
	}
}

// GetValue returns the current value for a field.
func (f *Form) GetValue(key string) string {
	return f.values[key]
}

// SetValue sets the value for a field.
func (f *Form) SetValue(key, value string) {
	f.values[key] = value
}

// Focus sets focus to the form.
func (f *Form) Focus(delegate func(p tview.Primitive)) {
	delegate(f.form)
}

// ShowForm displays a form and handles submit/cancel.
func (a *App) ShowForm(title string, fields []FormField, onSubmit func(map[string]string)) {
	onClose := func() {
		a.content.Pop()
		if view := a.getCurrentView(); view != nil {
			a.SetFocus(view.Primitive())
		}
	}

	form := NewForm(a, title, fields, func(values map[string]string) {
		onClose()
		if onSubmit != nil {
			onSubmit(values)
		}
	}, onClose)

	a.content.Push("form", form)
	a.SetFocus(form)
}
