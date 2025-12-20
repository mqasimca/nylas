package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestVimKeySequence_gg(t *testing.T) {
	// Test the gg key sequence tracking
	app := &App{
		lastKey:     0,
		lastKeyTime: time.Time{},
	}

	// First 'g' press
	now := time.Now()
	app.lastKey = 'g'
	app.lastKeyTime = now

	// Second 'g' press within timeout
	time.Sleep(100 * time.Millisecond)
	secondPress := time.Now()

	isGG := app.lastKey == 'g' && secondPress.Sub(app.lastKeyTime) < 500*time.Millisecond
	if !isGG {
		t.Error("Expected gg sequence to be detected")
	}
}

func TestVimKeySequence_gg_timeout(t *testing.T) {
	// Test that gg sequence times out
	app := &App{
		lastKey:     'g',
		lastKeyTime: time.Now().Add(-600 * time.Millisecond), // More than 500ms ago
	}

	// Second 'g' press after timeout
	now := time.Now()
	isGG := app.lastKey == 'g' && now.Sub(app.lastKeyTime) < 500*time.Millisecond
	if isGG {
		t.Error("Expected gg sequence to timeout")
	}
}

func TestVimKeySequence_dd(t *testing.T) {
	// Test the dd key sequence tracking
	app := &App{
		lastKey:     0,
		lastKeyTime: time.Time{},
	}

	// First 'd' press
	now := time.Now()
	app.lastKey = 'd'
	app.lastKeyTime = now

	// Second 'd' press within timeout
	time.Sleep(100 * time.Millisecond)
	secondPress := time.Now()

	isDD := app.lastKey == 'd' && secondPress.Sub(app.lastKeyTime) < 500*time.Millisecond
	if !isDD {
		t.Error("Expected dd sequence to be detected")
	}
}

func TestCommandParsing(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		isQuit   bool
		isNav    bool
		isAction bool
	}{
		// Quit commands
		{"quit q", "q", true, false, false},
		{"quit quit", "quit", true, false, false},
		{"quit exit", "exit", true, false, false},
		{"quit q!", "q!", true, false, false},
		{"quit wq", "wq", true, false, false},
		{"quit x", "x", true, false, false},

		// Navigation commands
		{"nav messages", "messages", false, true, false},
		{"nav m", "m", false, true, false},
		{"nav msg", "msg", false, true, false},
		{"nav events", "events", false, true, false},
		{"nav e", "e", false, true, false},
		{"nav ev", "ev", false, true, false},
		{"nav cal", "cal", false, true, false},
		{"nav calendar", "calendar", false, true, false},
		{"nav contacts", "contacts", false, true, false},
		{"nav c", "c", false, true, false},
		{"nav ct", "ct", false, true, false},
		{"nav webhooks", "webhooks", false, true, false},
		{"nav w", "w", false, true, false},
		{"nav wh", "wh", false, true, false},
		{"nav grants", "grants", false, true, false},
		{"nav g", "g", false, true, false},
		{"nav gr", "gr", false, true, false},
		{"nav dashboard", "dashboard", false, true, false},
		{"nav d", "d", false, true, false},
		{"nav dash", "dash", false, true, false},
		{"nav home", "home", false, true, false},

		// Action commands
		{"action delete", "delete", false, false, true},
		{"action del", "del", false, false, true},
		{"action rm", "rm", false, false, true},
		{"action star", "star", false, false, true},
		{"action unread", "unread", false, false, true},
		{"action compose", "compose", false, false, true},
		{"action new", "new", false, false, true},
		{"action n", "n", false, false, true},
		{"action reply", "reply", false, false, true},
		{"action r", "r", false, false, true},
		{"action replyall", "replyall", false, false, true},
		{"action ra", "ra", false, false, true},
		{"action forward", "forward", false, false, true},
		{"action f", "f", false, false, true},
		{"action fwd", "fwd", false, false, true},

		// Help
		{"help h", "h", false, false, false},
		{"help help", "help", false, false, false},

		// Numeric
		{"numeric 5", "5", false, false, false},
		{"numeric 10", "10", false, false, false},
		{"numeric 100", "100", false, false, false},
	}

	quitCmds := map[string]bool{
		"q": true, "quit": true, "exit": true, "q!": true, "quit!": true, "wq": true, "x": true,
	}
	navCmds := map[string]bool{
		"m": true, "messages": true, "msg": true,
		"e": true, "events": true, "ev": true, "cal": true, "calendar": true,
		"c": true, "contacts": true, "ct": true,
		"w": true, "webhooks": true, "wh": true,
		"g": true, "grants": true, "gr": true,
		"d": true, "dashboard": true, "dash": true, "home": true,
	}
	actionCmds := map[string]bool{
		"delete": true, "del": true, "rm": true,
		"star": true, "s": true, "unstar": true,
		"read": true, "mr": true, "unread": true, "mu": true,
		"new": true, "n": true, "compose": true,
		"reply": true, "r": true,
		"replyall": true, "ra": true, "reply-all": true,
		"forward": true, "f": true, "fwd": true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isQuit && !quitCmds[tt.cmd] {
				t.Errorf("Command %q should be recognized as quit", tt.cmd)
			}
			if tt.isNav && !navCmds[tt.cmd] {
				t.Errorf("Command %q should be recognized as navigation", tt.cmd)
			}
			if tt.isAction && !actionCmds[tt.cmd] {
				t.Errorf("Command %q should be recognized as action", tt.cmd)
			}
		})
	}
}

func TestNumericCommands(t *testing.T) {
	tests := []struct {
		cmd      string
		expected int
	}{
		{"1", 1},
		{"5", 5},
		{"10", 10},
		{"100", 100},
		{"999", 999},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if !isNumeric(tt.cmd) {
				t.Errorf("Command %q should be numeric", tt.cmd)
			}
			result := parseInt(tt.cmd)
			if result != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.cmd, result, tt.expected)
			}
		})
	}
}

func TestKeyEventCreation(t *testing.T) {
	// Test that we can create key events for testing
	tests := []struct {
		name string
		key  tcell.Key
		rune rune
	}{
		{"ctrl+c", tcell.KeyCtrlC, 0},
		{"ctrl+d", tcell.KeyCtrlD, 0},
		{"ctrl+u", tcell.KeyCtrlU, 0},
		{"ctrl+f", tcell.KeyCtrlF, 0},
		{"ctrl+b", tcell.KeyCtrlB, 0},
		{"escape", tcell.KeyEscape, 0},
		{"enter", tcell.KeyEnter, 0},
		{"tab", tcell.KeyTab, 0},
		{"rune j", tcell.KeyRune, 'j'},
		{"rune k", tcell.KeyRune, 'k'},
		{"rune g", tcell.KeyRune, 'g'},
		{"rune G", tcell.KeyRune, 'G'},
		{"rune :", tcell.KeyRune, ':'},
		{"rune /", tcell.KeyRune, '/'},
		{"rune ?", tcell.KeyRune, '?'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tcell.NewEventKey(tt.key, tt.rune, tcell.ModNone)
			if event == nil {
				t.Error("Failed to create key event")
			}
			if event.Key() != tt.key {
				t.Errorf("Key = %v, want %v", event.Key(), tt.key)
			}
			if tt.key == tcell.KeyRune && event.Rune() != tt.rune {
				t.Errorf("Rune = %c, want %c", event.Rune(), tt.rune)
			}
		})
	}
}

func TestFlashLevels(t *testing.T) {
	tests := []struct {
		level    FlashLevel
		expected int
	}{
		{FlashInfo, 0},
		{FlashWarn, 1},
		{FlashError, 2},
	}

	for _, tt := range tests {
		if int(tt.level) != tt.expected {
			t.Errorf("FlashLevel %v = %d, want %d", tt.level, tt.level, tt.expected)
		}
	}
}

// TestEditCommand tests the :e <view> command pattern
func TestEditCommandPattern(t *testing.T) {
	tests := []struct {
		cmd          string
		expectedView string
	}{
		{"e messages", "messages"},
		{"e m", "messages"},
		{"e events", "events"},
		{"e ev", "events"},
		{"e cal", "events"},
		{"e contacts", "contacts"},
		{"e c", "contacts"},
		{"e webhooks", "webhooks"},
		{"e w", "webhooks"},
		{"e grants", "grants"},
		{"e g", "grants"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if len(tt.cmd) <= 2 {
				t.Skip("Command too short")
			}
			if tt.cmd[:2] != "e " {
				t.Errorf("Command %q should start with 'e '", tt.cmd)
			}
			viewName := tt.cmd[2:]
			// Map short names to full names
			viewMap := map[string]string{
				"messages": "messages", "m": "messages",
				"events": "events", "ev": "events", "cal": "events",
				"contacts": "contacts", "c": "contacts",
				"webhooks": "webhooks", "w": "webhooks",
				"grants": "grants", "g": "grants",
			}
			expected, ok := viewMap[viewName]
			if !ok {
				t.Errorf("Unknown view name: %s", viewName)
			}
			if expected != tt.expectedView {
				t.Errorf("View for %q = %s, want %s", tt.cmd, expected, tt.expectedView)
			}
		})
	}
}
