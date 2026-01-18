package email

import (
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

// DisplayThreadListItem formats and prints a single thread in list view
func DisplayThreadListItem(t domain.Thread, showID bool) {
	status := " "
	if t.Unread {
		status = common.Cyan.Sprint("●")
	}

	star := " "
	if t.Starred {
		star = common.Yellow.Sprint("★")
	}

	attach := " "
	if t.HasAttachments {
		attach = "📎"
	}

	// Format participants
	participants := common.FormatParticipants(t.Participants)
	if len(participants) > 25 {
		participants = participants[:22] + "..."
	}

	subj := t.Subject
	if len(subj) > 35 {
		subj = subj[:32] + "..."
	}

	msgCount := fmt.Sprintf("(%d)", len(t.MessageIDs))
	dateStr := common.FormatTimeAgo(t.LatestMessageRecvDate)

	fmt.Printf("%s %s %s %-25s %-35s %-5s %s\n",
		status, star, attach, participants, subj, msgCount, common.Dim.Sprint(dateStr))

	if showID {
		_, _ = common.Dim.Printf("      ID: %s\n", t.ID)
	}
}
