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

// WebhookFormMode indicates if we're creating or editing.
type WebhookFormMode int

const (
	WebhookFormCreate WebhookFormMode = iota
	WebhookFormEdit
)

// WebhookForm provides a form for creating/editing webhooks.
type WebhookForm struct {
	*tview.Flex
	app      *App
	form     *tview.Form
	mode     WebhookFormMode
	webhook  *domain.Webhook // nil for create, populated for edit
	onSubmit func(*domain.Webhook)
	onCancel func()

	// Form field values
	webhookURL   string
	description  string
	triggerTypes []string
	notifyEmails string
	status       string
}

// Common trigger type presets for easier selection.
var triggerPresets = map[string][]string{
	"All Messages": {
		domain.TriggerMessageCreated,
		domain.TriggerMessageUpdated,
	},
	"Message Tracking": {
		domain.TriggerMessageOpened,
		domain.TriggerMessageLinkClicked,
		domain.TriggerMessageBounceDetected,
	},
	"All Events": {
		domain.TriggerEventCreated,
		domain.TriggerEventUpdated,
		domain.TriggerEventDeleted,
	},
	"All Contacts": {
		domain.TriggerContactCreated,
		domain.TriggerContactUpdated,
		domain.TriggerContactDeleted,
	},
	"Grant Changes": {
		domain.TriggerGrantCreated,
		domain.TriggerGrantUpdated,
		domain.TriggerGrantDeleted,
		domain.TriggerGrantExpired,
	},
}

// NewWebhookForm creates a new webhook form.
func NewWebhookForm(app *App, webhook *domain.Webhook, onSubmit func(*domain.Webhook), onCancel func()) *WebhookForm {
	mode := WebhookFormCreate
	if webhook != nil {
		mode = WebhookFormEdit
	}

	f := &WebhookForm{
		Flex:         tview.NewFlex(),
		app:          app,
		mode:         mode,
		webhook:      webhook,
		onSubmit:     onSubmit,
		onCancel:     onCancel,
		triggerTypes: []string{},
		status:       "active",
	}

	// Populate from existing webhook if editing
	if webhook != nil {
		f.webhookURL = webhook.WebhookURL
		f.description = webhook.Description
		f.triggerTypes = webhook.TriggerTypes
		f.notifyEmails = strings.Join(webhook.NotificationEmailAddresses, ", ")
		f.status = webhook.Status
	}

	f.init()
	return f
}

func (f *WebhookForm) init() {
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

	title := "New Webhook"
	if f.mode == WebhookFormEdit {
		title = "Edit Webhook"
	}
	f.form.SetTitle(fmt.Sprintf(" %s ", title))
	f.form.SetTitleColor(styles.TitleFg)

	// Add form fields
	f.form.AddInputField("Webhook URL", f.webhookURL, 50, nil, func(text string) {
		f.webhookURL = text
	})

	f.form.AddInputField("Description", f.description, 50, nil, func(text string) {
		f.description = text
	})

	// Trigger type preset selection
	presetOptions := []string{"Custom", "All Messages", "Message Tracking", "All Events", "All Contacts", "Grant Changes"}
	f.form.AddDropDown("Trigger Preset", presetOptions, 0, func(option string, _ int) {
		if triggers, ok := triggerPresets[option]; ok {
			f.triggerTypes = triggers
		}
	})

	// Manual trigger types input (comma-separated)
	triggersStr := strings.Join(f.triggerTypes, ", ")
	f.form.AddInputField("Triggers (comma-sep)", triggersStr, 50, nil, func(text string) {
		parts := strings.Split(text, ",")
		f.triggerTypes = make([]string, 0)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				f.triggerTypes = append(f.triggerTypes, p)
			}
		}
	})

	f.form.AddInputField("Notify Emails (comma-sep)", f.notifyEmails, 50, nil, func(text string) {
		f.notifyEmails = text
	})

	// Status (only for edit mode)
	if f.mode == WebhookFormEdit {
		statusOptions := []string{"active", "inactive"}
		statusIndex := 0
		for i, s := range statusOptions {
			if s == f.status {
				statusIndex = i
				break
			}
		}
		f.form.AddDropDown("Status", statusOptions, statusIndex, func(option string, _ int) {
			f.status = option
		})
	}

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
		AddItem(f.form, 70, 0, true).
		AddItem(nil, 0, 1, false), 0, 3, true)
	f.AddItem(nil, 0, 1, false)
}

func (f *WebhookForm) handleInput(event *tcell.EventKey) *tcell.EventKey {
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

func (f *WebhookForm) validate() []string {
	var errors []string

	if strings.TrimSpace(f.webhookURL) == "" {
		errors = append(errors, "Webhook URL is required")
	} else if !strings.HasPrefix(f.webhookURL, "http://") && !strings.HasPrefix(f.webhookURL, "https://") {
		errors = append(errors, "Webhook URL must start with http:// or https://")
	}

	if len(f.triggerTypes) == 0 {
		errors = append(errors, "At least one trigger type is required")
	}

	return errors
}

func (f *WebhookForm) submit() {
	// Validate
	errors := f.validate()
	if len(errors) > 0 {
		f.app.Flash(FlashError, "%s", strings.Join(errors, "; "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse notification emails
	var notifyEmails []string
	if f.notifyEmails != "" {
		for _, email := range strings.Split(f.notifyEmails, ",") {
			email = strings.TrimSpace(email)
			if email != "" {
				notifyEmails = append(notifyEmails, email)
			}
		}
	}

	var resultWebhook *domain.Webhook
	var err error

	if f.mode == WebhookFormCreate {
		// Create new webhook
		req := &domain.CreateWebhookRequest{
			WebhookURL:                 strings.TrimSpace(f.webhookURL),
			Description:                strings.TrimSpace(f.description),
			TriggerTypes:               f.triggerTypes,
			NotificationEmailAddresses: notifyEmails,
		}
		resultWebhook, err = f.app.config.Client.CreateWebhook(ctx, req)
	} else {
		// Update existing webhook
		req := &domain.UpdateWebhookRequest{
			WebhookURL:                 strings.TrimSpace(f.webhookURL),
			Description:                strings.TrimSpace(f.description),
			TriggerTypes:               f.triggerTypes,
			NotificationEmailAddresses: notifyEmails,
			Status:                     f.status,
		}
		resultWebhook, err = f.app.config.Client.UpdateWebhook(ctx, f.webhook.ID, req)
	}

	if err != nil {
		f.app.Flash(FlashError, "Failed to save webhook: %v", err)
		return
	}

	if f.mode == WebhookFormCreate {
		f.app.Flash(FlashInfo, "Webhook created: %s", resultWebhook.ID)
	} else {
		f.app.Flash(FlashInfo, "Webhook updated: %s", resultWebhook.ID)
	}

	if f.onSubmit != nil {
		f.onSubmit(resultWebhook)
	}
}

func (f *WebhookForm) cancel() {
	if f.onCancel != nil {
		f.onCancel()
	}
}

// Focus sets focus to the form.
func (f *WebhookForm) Focus(delegate func(p tview.Primitive)) {
	delegate(f.form)
}

// ShowWebhookForm displays a webhook form for create/edit.
func (a *App) ShowWebhookForm(webhook *domain.Webhook, onSave func(*domain.Webhook)) {
	onClose := func() {
		a.content.Pop()
		if view := a.getCurrentView(); view != nil {
			a.SetFocus(view.Primitive())
		}
	}

	form := NewWebhookForm(a, webhook, func(savedWebhook *domain.Webhook) {
		onClose()
		if onSave != nil {
			onSave(savedWebhook)
		}
	}, onClose)

	a.content.Push("webhook-form", form)
	a.SetFocus(form)
}

// DeleteWebhook shows a confirmation dialog and deletes a webhook.
func (a *App) DeleteWebhook(webhook *domain.Webhook, onDelete func()) {
	desc := webhook.Description
	if desc == "" {
		desc = webhook.ID
	}
	a.ShowConfirmDialog("Delete Webhook", fmt.Sprintf("Delete webhook '%s'?", desc), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := a.config.Client.DeleteWebhook(ctx, webhook.ID)
		if err != nil {
			a.Flash(FlashError, "Failed to delete webhook: %v", err)
			return
		}

		a.Flash(FlashInfo, "Webhook deleted")
		if onDelete != nil {
			onDelete()
		}
	})
}
