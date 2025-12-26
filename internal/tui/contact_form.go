package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/rivo/tview"
)

// ContactFormMode indicates if we're creating or editing.
type ContactFormMode int

const (
	ContactFormCreate ContactFormMode = iota
	ContactFormEdit
)

// ContactForm provides a form for creating/editing contacts.
type ContactForm struct {
	*tview.Flex
	app      *App
	form     *tview.Form
	mode     ContactFormMode
	contact  *domain.Contact // nil for create, populated for edit
	onSubmit func(*domain.Contact)
	onCancel func()

	// Form field values
	givenName   string
	surname     string
	email       string
	emailType   string
	phone       string
	phoneType   string
	companyName string
	jobTitle    string
	notes       string
}

// NewContactForm creates a new contact form.
func NewContactForm(app *App, contact *domain.Contact, onSubmit func(*domain.Contact), onCancel func()) *ContactForm {
	mode := ContactFormCreate
	if contact != nil {
		mode = ContactFormEdit
	}

	f := &ContactForm{
		Flex:     tview.NewFlex(),
		app:      app,
		mode:     mode,
		contact:  contact,
		onSubmit: onSubmit,
		onCancel: onCancel,
	}

	// Populate from existing contact if editing
	if contact != nil {
		f.givenName = contact.GivenName
		f.surname = contact.Surname
		f.companyName = contact.CompanyName
		f.jobTitle = contact.JobTitle
		f.notes = contact.Notes

		if len(contact.Emails) > 0 {
			f.email = contact.Emails[0].Email
			f.emailType = contact.Emails[0].Type
		}
		if len(contact.PhoneNumbers) > 0 {
			f.phone = contact.PhoneNumbers[0].Number
			f.phoneType = contact.PhoneNumbers[0].Type
		}
	} else {
		f.emailType = "work"
		f.phoneType = "mobile"
	}

	f.init()
	return f
}

func (f *ContactForm) init() {
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

	title := "New Contact"
	if f.mode == ContactFormEdit {
		title = "Edit Contact"
	}
	f.form.SetTitle(fmt.Sprintf(" %s ", title))
	f.form.SetTitleColor(styles.TitleFg)

	// Add form fields
	f.form.AddInputField("First Name", f.givenName, 30, nil, func(text string) {
		f.givenName = text
	})

	f.form.AddInputField("Last Name", f.surname, 30, nil, func(text string) {
		f.surname = text
	})

	f.form.AddInputField("Email", f.email, 40, nil, func(text string) {
		f.email = text
	})

	emailTypes := []string{"work", "home", "other"}
	emailTypeIndex := 0
	for i, t := range emailTypes {
		if t == f.emailType {
			emailTypeIndex = i
			break
		}
	}
	f.form.AddDropDown("Email Type", emailTypes, emailTypeIndex, func(option string, _ int) {
		f.emailType = option
	})

	f.form.AddInputField("Phone", f.phone, 20, nil, func(text string) {
		f.phone = text
	})

	phoneTypes := []string{"mobile", "work", "home", "other"}
	phoneTypeIndex := 0
	for i, t := range phoneTypes {
		if t == f.phoneType {
			phoneTypeIndex = i
			break
		}
	}
	f.form.AddDropDown("Phone Type", phoneTypes, phoneTypeIndex, func(option string, _ int) {
		f.phoneType = option
	})

	f.form.AddInputField("Company", f.companyName, 30, nil, func(text string) {
		f.companyName = text
	})

	f.form.AddInputField("Job Title", f.jobTitle, 30, nil, func(text string) {
		f.jobTitle = text
	})

	f.form.AddTextArea("Notes", f.notes, 40, 3, 0, func(text string) {
		f.notes = text
	})

	// Add buttons
	f.form.AddButton("Save", f.submit)
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

func (f *ContactForm) handleInput(event *tcell.EventKey) *tcell.EventKey {
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

func (f *ContactForm) validate() []string {
	var errors []string

	// At least one of given name, surname, or email is required
	hasName := strings.TrimSpace(f.givenName) != "" || strings.TrimSpace(f.surname) != ""
	hasEmail := strings.TrimSpace(f.email) != ""

	if !hasName && !hasEmail {
		errors = append(errors, "At least a name or email is required")
	}

	// Validate email format if provided
	if f.email != "" && !strings.Contains(f.email, "@") {
		errors = append(errors, "Email must be a valid email address")
	}

	return errors
}

func (f *ContactForm) submit() {
	// Validate
	errors := f.validate()
	if len(errors) > 0 {
		f.app.Flash(FlashError, "%s", strings.Join(errors, "; "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var resultContact *domain.Contact
	var err error

	if f.mode == ContactFormCreate {
		// Create new contact
		req := &domain.CreateContactRequest{
			GivenName:   strings.TrimSpace(f.givenName),
			Surname:     strings.TrimSpace(f.surname),
			CompanyName: strings.TrimSpace(f.companyName),
			JobTitle:    strings.TrimSpace(f.jobTitle),
			Notes:       strings.TrimSpace(f.notes),
		}

		if f.email != "" {
			req.Emails = []domain.ContactEmail{{
				Email: strings.TrimSpace(f.email),
				Type:  f.emailType,
			}}
		}

		if f.phone != "" {
			req.PhoneNumbers = []domain.ContactPhone{{
				Number: strings.TrimSpace(f.phone),
				Type:   f.phoneType,
			}}
		}

		resultContact, err = f.app.config.Client.CreateContact(ctx, f.app.config.GrantID, req)
	} else {
		// Update existing contact
		givenName := strings.TrimSpace(f.givenName)
		surname := strings.TrimSpace(f.surname)
		companyName := strings.TrimSpace(f.companyName)
		jobTitle := strings.TrimSpace(f.jobTitle)
		notes := strings.TrimSpace(f.notes)

		req := &domain.UpdateContactRequest{
			GivenName:   &givenName,
			Surname:     &surname,
			CompanyName: &companyName,
			JobTitle:    &jobTitle,
			Notes:       &notes,
		}

		if f.email != "" {
			req.Emails = []domain.ContactEmail{{
				Email: strings.TrimSpace(f.email),
				Type:  f.emailType,
			}}
		}

		if f.phone != "" {
			req.PhoneNumbers = []domain.ContactPhone{{
				Number: strings.TrimSpace(f.phone),
				Type:   f.phoneType,
			}}
		}

		resultContact, err = f.app.config.Client.UpdateContact(ctx, f.app.config.GrantID, f.contact.ID, req)
	}

	if err != nil {
		f.app.Flash(FlashError, "Failed to save contact: %v", err)
		return
	}

	if f.mode == ContactFormCreate {
		f.app.Flash(FlashInfo, "Contact created: %s", resultContact.DisplayName())
	} else {
		f.app.Flash(FlashInfo, "Contact updated: %s", resultContact.DisplayName())
	}

	if f.onSubmit != nil {
		f.onSubmit(resultContact)
	}
}

func (f *ContactForm) cancel() {
	if f.onCancel != nil {
		f.onCancel()
	}
}

// Focus sets focus to the form.
func (f *ContactForm) Focus(delegate func(p tview.Primitive)) {
	delegate(f.form)
}

// ShowContactForm displays a contact form for create/edit.
func (a *App) ShowContactForm(contact *domain.Contact, onSave func(*domain.Contact)) {
	onClose := func() {
		a.content.Pop()
		if view := a.getCurrentView(); view != nil {
			a.SetFocus(view.Primitive())
		}
	}

	form := NewContactForm(a, contact, func(savedContact *domain.Contact) {
		onClose()
		if onSave != nil {
			onSave(savedContact)
		}
	}, onClose)

	a.content.Push("contact-form", form)
	a.SetFocus(form)
}

// DeleteContact shows a confirmation dialog and deletes a contact.
func (a *App) DeleteContact(contact *domain.Contact, onDelete func()) {
	a.ShowConfirmDialog("Delete Contact", fmt.Sprintf("Delete contact '%s'?", contact.DisplayName()), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := a.config.Client.DeleteContact(ctx, a.config.GrantID, contact.ID)
		if err != nil {
			a.Flash(FlashError, "Failed to delete contact: %v", err)
			return
		}

		a.Flash(FlashInfo, "Contact deleted: %s", contact.DisplayName())
		if onDelete != nil {
			onDelete()
		}
	})
}
