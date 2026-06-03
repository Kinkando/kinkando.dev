package gemini

import "testing"

func TestResolvePersona(t *testing.T) {
	cases := []struct {
		name    string
		history []Message
		userMsg string
		want    persona
	}{
		// --- name mention (English) ---
		{"mint name", nil, "Mint, log 200 THB coffee", personaMint},
		{"kaito name", nil, "kaito add a card", personaKaito},
		{"tensei name", nil, "Tensei, how did I sleep last night?", personaTensei},
		{"KAITO uppercase", nil, "KAITO, create a card", personaKaito},
		{"MINT uppercase", nil, "MINT how much did I spend?", personaMint},
		{"TENSEI uppercase", nil, "TENSEI log my workout", personaTensei},
		// Name must be a whole word (no false match inside other words).
		{"no match: terminate", nil, "terminate the process", personaAether},

		// --- name mention (Thai) ---
		{"thai kaito name", nil, "ไคโตะ ช่วยเพิ่ม card หน่อย", personaKaito},
		{"thai mint name (มิ้นต์)", nil, "มิ้นต์ บันทึกค่ากาแฟ 60 บาทหน่อย", personaMint},
		{"thai mint alt (มินต์)", nil, "มินต์ สรุปรายจ่ายเดือนนี้", personaMint},
		{"thai kaito mid-sentence", nil, "อยากให้ ไคโตะ ย้าย card ไปคอลัมน์ Done", personaKaito},
		{"thai tensei name (เทนเซ)", nil, "เทนเซ ดูประวัติการนอนหน่อย", personaTensei},
		{"thai tensei alt (เท็นเซ)", nil, "เท็นเซ บันทึกการออกกำลังกาย", personaTensei},

		// --- keyword fallback (English) ---
		{"keyword: spend → Mint", nil, "how much did I spend in May?", personaMint},
		{"keyword: expense → Mint", nil, "add a new expense", personaMint},
		{"keyword: income → Mint", nil, "record my income for today", personaMint},
		{"keyword: card → Kaito", nil, "move this card to done", personaKaito},
		{"keyword: board → Kaito", nil, "show me the board", personaKaito},
		{"keyword: task → Kaito", nil, "I have a new task", personaKaito},
		// workout
		{"keyword: workout → Tensei", nil, "log my workout for today", personaTensei},
		{"keyword: running → Tensei", nil, "I went running this morning", personaTensei},
		{"keyword: gym → Tensei", nil, "I'm going to the gym", personaTensei},
		{"keyword: session → Tensei", nil, "start a new session", personaTensei},
		// sleep
		{"keyword: sleep → Tensei", nil, "how did I sleep this week?", personaTensei},
		{"keyword: bedtime → Tensei", nil, "log my bedtime", personaTensei},
		{"keyword: recovery → Tensei", nil, "check my recovery this week", personaTensei},
		{"keyword: sleep score → Tensei", nil, "my sleep score was 72 last night", personaTensei},
		// food & nutrition
		{"keyword: food → Tensei", nil, "log my food for today", personaTensei},
		{"keyword: meal → Tensei", nil, "I had a big meal", personaTensei},
		{"keyword: calories → Tensei", nil, "how many calories did I eat?", personaTensei},
		{"keyword: protein → Tensei", nil, "I need more protein today", personaTensei},
		{"keyword: breakfast → Tensei", nil, "log my breakfast", personaTensei},
		{"keyword: macros → Tensei", nil, "check my macros for today", personaTensei},

		// --- keyword fallback (Thai) ---
		{"thai keyword: รายจ่าย → Mint", nil, "ดูรายจ่ายเดือนนี้หน่อย", personaMint},
		{"thai keyword: รายรับ → Mint", nil, "บันทึกรายรับวันนี้", personaMint},
		{"thai keyword: ค่าใช้จ่าย → Mint", nil, "สรุปค่าใช้จ่าย", personaMint},
		{"thai keyword: งบประมาณ → Mint", nil, "เช็คงบประมาณหน่อย", personaMint},
		{"thai keyword: เงินเดือน → Mint", nil, "บันทึกเงินเดือนเดือนนี้", personaMint},
		{"thai keyword: การ์ด → Kaito", nil, "เพิ่มการ์ดใหม่หน่อย", personaKaito},
		{"thai keyword: บอร์ด → Kaito", nil, "ดูบอร์ดทั้งหมด", personaKaito},
		{"thai keyword: แคนบัน → Kaito", nil, "อัปเดตแคนบัน", personaKaito},
		{"thai keyword: คอลัมน์ → Kaito", nil, "ย้ายไปคอลัมน์ Done", personaKaito},
		// Thai workout
		{"thai keyword: ออกกำลังกาย → Tensei", nil, "ออกกำลังกายวันนี้", personaTensei},
		{"thai keyword: ฟิตเนส → Tensei", nil, "ไปฟิตเนส", personaTensei},
		{"thai keyword: วิ่ง → Tensei", nil, "วิ่งเช้านี้ 5 กม.", personaTensei},
		// Thai sleep
		{"thai keyword: นอนหลับ → Tensei", nil, "ดูสถิตินอนหลับ", personaTensei},
		{"thai keyword: การนอน → Tensei", nil, "สรุปการนอนสัปดาห์นี้", personaTensei},
		{"thai keyword: คะแนนการนอน → Tensei", nil, "คะแนนการนอนคืนนี้ 80", personaTensei},
		{"thai keyword: พักผ่อน → Tensei", nil, "พักผ่อนไม่พอวันนี้", personaTensei},
		// Thai food
		{"thai keyword: อาหาร → Tensei", nil, "บันทึกอาหารกลางวัน", personaTensei},
		{"thai keyword: แคลอรี่ → Tensei", nil, "กินแคลอรี่ไปเท่าไหร่", personaTensei},
		{"thai keyword: โปรตีน → Tensei", nil, "โปรตีนวันนี้พอไหม", personaTensei},
		{"thai keyword: อาหารเช้า → Tensei", nil, "บันทึกอาหารเช้า", personaTensei},

		// --- Aether default ---
		{"aether: greeting", nil, "hello, who are you?", personaAether},
		{"aether: empty", nil, "", personaAether},
		{"aether: thai greeting", nil, "สวัสดี คุณคือใคร", personaAether},

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
		{
			name: "sticky: Tensei from history (workout)",
			history: []Message{
				{Role: "user", Text: "Tensei, start a workout session"},
				{Role: "model", Text: "Session started."},
			},
			userMsg: "log 3 sets of 10 reps",
			want:    personaTensei,
		},
		{
			name: "sticky: Tensei from history (sleep)",
			history: []Message{
				{Role: "user", Text: "Tensei, log my sleep"},
				{Role: "model", Text: "Sleep logged."},
			},
			userMsg: "score was 78",
			want:    personaTensei,
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
		{
			name: "sticky: Thai name in history",
			history: []Message{
				{Role: "user", Text: "มิ้นต์ ดูรายจ่ายเดือนก่อนหน่อย"},
				{Role: "model", Text: "นี่คือรายจ่ายของคุณ"},
			},
			userMsg: "ขอบคุณ",
			want:    personaMint,
		},
		{
			name: "sticky: Thai Tensei from history",
			history: []Message{
				{Role: "user", Text: "เทนเซ บันทึกการนอนคืนนี้"},
				{Role: "model", Text: "บันทึกแล้ว"},
			},
			userMsg: "ขอบคุณ",
			want:    personaTensei,
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
		{
			name: "thai name overrides history",
			history: []Message{
				{Role: "user", Text: "มิ้นต์ สรุปรายรับ"},
			},
			userMsg: "ไคโตะ เพิ่มงานใหม่",
			want:    personaKaito,
		},
		{
			name: "tensei overrides finance history",
			history: []Message{
				{Role: "user", Text: "Mint, show my expenses"},
			},
			userMsg: "Tensei how did I sleep?",
			want:    personaTensei,
		},

		// --- Kusuri: name mention (English) ---
		{"kusuri name", nil, "Kusuri, did I take my pills today?", personaKusuri},
		{"KUSURI uppercase", nil, "KUSURI how much stock do I have left?", personaKusuri},

		// --- Kusuri: name mention (Thai) ---
		{"thai kusuri name", nil, "คุสุริ บันทึกการกินยา", personaKusuri},

		// --- Kusuri: keyword fallback (English) ---
		{"keyword: medicine → Kusuri", nil, "how much medicine do I have left?", personaKusuri},
		{"keyword: pill → Kusuri", nil, "I took a pill this morning", personaKusuri},
		{"keyword: pills → Kusuri", nil, "remind me to take my pills", personaKusuri},
		{"keyword: tablet → Kusuri", nil, "I only have 3 tablets left", personaKusuri},
		{"keyword: capsule → Kusuri", nil, "take one capsule after meals", personaKusuri},
		{"keyword: dose → Kusuri", nil, "what's my dose?", personaKusuri},
		{"keyword: dosage → Kusuri", nil, "log my dosage", personaKusuri},
		{"keyword: prescription → Kusuri", nil, "check my prescription refill", personaKusuri},
		{"keyword: refill → Kusuri", nil, "I need a refill", personaKusuri},
		{"keyword: restock → Kusuri", nil, "I restocked my medication", personaKusuri},
		{"keyword: medication → Kusuri", nil, "list all my medication", personaKusuri},
		{"keyword: pharmacy → Kusuri", nil, "I went to the pharmacy", personaKusuri},

		// --- Kusuri: keyword fallback (Thai) ---
		{"thai keyword: กินยา → Kusuri", nil, "บันทึกกินยาเช้านี้", personaKusuri},
		{"thai keyword: ยาเม็ด → Kusuri", nil, "เหลือยาเม็ดกี่เม็ด", personaKusuri},
		{"thai keyword: สต็อกยา → Kusuri", nil, "เช็คสต็อกยา", personaKusuri},
		{"thai keyword: ยาที่เหลือ → Kusuri", nil, "ยาที่เหลืออยู่เท่าไหร่", personaKusuri},
		{"thai keyword: เม็ดยา → Kusuri", nil, "เม็ดยาหมดแล้ว", personaKusuri},
		{"thai keyword: โดสยา → Kusuri", nil, "บันทึกโดสยา", personaKusuri},

		// --- Kusuri: sticky via history ---
		{
			name: "sticky: Kusuri from history",
			history: []Message{
				{Role: "user", Text: "Kusuri, show my medicines"},
				{Role: "model", Text: "Here are your medicines."},
			},
			userMsg: "log that I just took one",
			want:    personaKusuri,
		},
		{
			name: "sticky: Thai Kusuri from history",
			history: []Message{
				{Role: "user", Text: "คุสุริ ดูสต็อกยา"},
				{Role: "model", Text: "สต็อกยาของคุณมีดังนี้"},
			},
			userMsg: "ขอบคุณ",
			want:    personaKusuri,
		},

		// --- Guard: no false-positive on bare ยา substring ---
		{"no mis-route: อยากกินข้าว", nil, "อยากกินข้าว", personaAether},
		// "อยาก" contains the sequence ยา as a substring; must not trigger Kusuri.
		{"no mis-route: อยากดูหนัง", nil, "ฉันอยากดูหนัง", personaAether},
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
