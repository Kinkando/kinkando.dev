package gemini

import "testing"

func TestResolvePersona(t *testing.T) {
	cases := []struct {
		name    string
		history []Message
		userMsg string
		want    persona
	}{
		// --- name mention ---
		{"mint name", nil, "Mint, log 200 THB coffee", personaMint},
		{"kaito name", nil, "kaito add a card", personaKaito},
		{"KAITO uppercase", nil, "KAITO, create a card", personaKaito},
		{"MINT uppercase", nil, "MINT how much did I spend?", personaMint},
		// Name must be a whole word (no false match inside other words).
		{"no match: terminate", nil, "terminate the process", personaAether},

		// --- keyword fallback ---
		{"keyword: spend → Mint", nil, "how much did I spend in May?", personaMint},
		{"keyword: expense → Mint", nil, "add a new expense", personaMint},
		{"keyword: income → Mint", nil, "record my income for today", personaMint},
		{"keyword: card → Kaito", nil, "move this card to done", personaKaito},
		{"keyword: board → Kaito", nil, "show me the board", personaKaito},
		{"keyword: task → Kaito", nil, "I have a new task", personaKaito},

		// --- Aether default ---
		{"aether: greeting", nil, "hello, who are you?", personaAether},
		{"aether: empty", nil, "", personaAether},

		// --- sticky via history ---
		{
			name: "sticky: Mint from history",
			history: []Message{
				{Role: "user", Text: "Mint, what's my balance?"},
				{Role: "model", Text: "Your balance is 500 THB."},
			},
			userMsg: "yes do it",
			want:    personaMint,
		},
		{
			name: "sticky: Kaito from history",
			history: []Message{
				{Role: "user", Text: "Kaito, show me the board"},
				{Role: "model", Text: "Here is your board."},
			},
			userMsg: "ok",
			want:    personaKaito,
		},
		// Model turns are skipped during sticky scan.
		{
			name: "sticky: skips model turns",
			history: []Message{
				{Role: "user", Text: "Mint, list my expenses"},
				{Role: "model", Text: "Here are your expenses."},
				{Role: "model", Text: "Anything else?"},
			},
			userMsg: "thanks",
			want:    personaMint,
		},

		// --- current message overrides history ---
		{
			name: "override history with name",
			history: []Message{
				{Role: "user", Text: "Mint, show spending"},
			},
			userMsg: "Kaito add a task",
			want:    personaKaito,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolvePersona(tc.history, tc.userMsg)
			if got != tc.want {
				t.Errorf("resolvePersona(history, %q) = %v, want %v", tc.userMsg, got, tc.want)
			}
		})
	}
}
