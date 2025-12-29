package models

import (
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// ComposeMode represents the mode of composition
type ComposeMode int

const (
	ComposeModeNew ComposeMode = iota
	ComposeModeReply
	ComposeModeReplyAll
	ComposeModeForward
	ComposeModeDraft
)

// ComposeData contains the data needed to initialize the compose screen
type ComposeData struct {
	Mode    ComposeMode
	Message *domain.Message
	Draft   *domain.Draft
}

// SaveStatus represents the draft save status
type SaveStatus int

const (
	SaveStatusNone SaveStatus = iota
	SaveStatusSaving
	SaveStatusSaved
	SaveStatusError
	SaveStatusUnsaved
)

// Compose is the email composition screen
type Compose struct {
	global *state.GlobalState
	theme  *styles.Theme

	// Compose context
	mode       ComposeMode
	replyToMsg *domain.Message
	draftID    string

	// Form fields (Bubbles components)
	toInput      textinput.Model
	ccInput      textinput.Model
	bccInput     textinput.Model
	subjectInput textinput.Model
	bodyInput    textarea.Model

	// Focus management
	focusIndex int // 0=to, 1=cc, 2=bcc, 3=subject, 4=body

	// State
	sending          bool
	savingDraft      bool
	validationErrors map[string]string
	showCc           bool
	showBcc          bool

	// Autosave
	lastSavedHash    string
	isDirty          bool
	lastSaveTime     time.Time
	saveStatus       SaveStatus
	autosaveEnabled  bool
	autosaveInterval time.Duration

	// Window dimensions
	width  int
	height int

	// Cursor repositioning (workaround for bubbles v0.21.0 viewport bug)
	needsCursorReposition bool
	hasReceivedWindowSize bool
}

// AutosaveTickMsg is sent by the autosave timer
type AutosaveTickMsg time.Time

// DraftSavedMsg is sent when a draft save completes
type DraftSavedMsg struct {
	draftID string
	err     error
}

// MessageSentMsg is sent when a message is successfully sent
type MessageSentMsg struct {
	message *domain.Message
}

// SendErrorMsg is sent when sending fails
type SendErrorMsg struct {
	err error
}

// NewCompose creates a new compose screen
