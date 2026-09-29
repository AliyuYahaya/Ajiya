package cli

import (
	"encoding/json"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
	"github.com/AliyuYahaya/Ajiya/internal/view"
)

// The JSON shapes live in internal/view; these names keep the CLI code short.
type (
	jsonStatus = view.Status
	jsonTicket = view.Ticket
)

var stateNames = view.StateNames

func toJSON(t *plan.Ticket) jsonTicket { return view.TicketOf(t) }

func writeJSON(e *env, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = e.stdout.Write(append(data, '\n'))
	return err
}
