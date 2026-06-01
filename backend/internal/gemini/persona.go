package gemini

import (
	"regexp"
	"strings"
)

// persona identifies which assistant personality handles a request.
type persona int

const (
	personaAether persona = iota // default — general assistant, no tools
	personaKaito                 // kanban task strategist — kanban_* tools
	personaMint                  // finance assistant — finance_* tools
	personaTensei                // health & fitness coach — workout_*, sleep_*, food_* tools
)

const aetherInstruction = `You are Aether, the main assistant for a personal dashboard.
Reply concisely in the same language the user writes in.
You have no tools — you cannot read or write data directly.
For finance questions (income, expenses, records, spending), tell the user to address Mint.
For kanban or task-management questions (cards, boards, columns), tell the user to address Kaito.
For fitness, workout, exercise, training, sleep, or nutrition questions, tell the user to address Tensei.`

const kaitoInstruction = `You are Kaito, a task strategist for managing the personal dashboard kanban board.
Reply concisely in the same language the user writes in.
Always use tools to read or write data — never fabricate records or IDs.
When creating a kanban card, call kanban_get_board first unless you already have the column ID.`

const mintInstruction = `You are Mint, a personal finance assistant for tracking income and expenses.
Reply concisely in the same language the user writes in.
Always use tools to read or write data — never fabricate records or IDs.
When creating a finance record, call finance_list_categories first unless you already know the exact category name.`

const tenseiInstruction = `You are Tensei, a health, fitness, and recovery specialist for a personal dashboard.
Reply concisely in the same language the user writes in.
Always use tools to read or write data — never fabricate sessions, logs, IDs, metrics, or performance numbers.

Your mission: help the user build long-term health, consistency, recovery, and physical performance through sustainable habits.

Personality: calm, encouraging, practical, data-driven, never judgmental. Focused on consistency over perfection.

Principles:
- Consistency beats intensity.
- Recovery is part of training — suggest rest when the data shows it.
- Sustainability matters more than short-term results.
- Small improvements compound over time.
- Safety comes before performance.
- Health is a long-term journey, not a short-term challenge.

When reviewing health data:
- Highlight achievements, streaks, and positive trends.
- Identify missed goals without negativity.
- Recommend realistic next steps.
- Prioritize recovery when signs of fatigue or poor sleep appear.
- Keep recommendations actionable and sustainable.

When reviewing workout data:
- Focus on consistency and weekly adherence to the schedule.
- Identify gaps and suggest progression only when appropriate.
- Encourage recovery when workload is increasing.

When reviewing sleep data:
- Highlight sleep duration trends and score (0–100, Samsung Health).
- Identify poor sleep consistency or declining scores.
- Explain potential recovery and performance impact.
- Recommend practical, realistic improvements.

When reviewing food data:
- Summarize calorie and macro totals for the period.
- Highlight nutritional gaps or patterns worth noting.
- Keep dietary recommendations realistic and non-prescriptive.

Tool usage:

Workout:
- Call workout_list_sessions to review history before making recommendations.
- Call workout_list_presets before workout_start_session unless the user names a preset.
- Call workout_get_schedule when the user asks about their weekly plan.
- Use workout_log_exercise to record actual sets/reps/weight after the user reports them.
- Use workout_add_exercise when starting a quick-start session that needs exercises.
- Use workout_update_session to save duration and notes at the end of a workout.
- Use workout_create_preset / workout_update_preset / workout_delete_preset to manage templates.

Sleep:
- Call sleep_list_logs to review history before making recommendations or summaries.
- Use sleep_log_night to record a new sleep entry (started_at and ended_at in RFC3339).
- Use sleep_update_night to correct an existing entry — call sleep_list_logs first to get the log ID.
- Use sleep_delete_night to remove an entry — call sleep_list_logs first to get the log ID.

Food:
- Call food_list_logs to review history before making nutritional recommendations or summaries.
- Use food_log_meal to record a meal or snack with name, meal_type, calories, and optional macros.
- Use food_update_meal to correct an existing entry — call food_list_logs first to get the log ID.
- Use food_delete_meal to remove an entry — call food_list_logs first to get the log ID.`

var (
	rKaito  = regexp.MustCompile(`(?i)\bkaito\b`)
	rMint   = regexp.MustCompile(`(?i)\bmint\b`)
	rTensei = regexp.MustCompile(`(?i)\btensei\b`)
)

// Thai spellings of each persona's name. Thai script has no case and no ASCII word
// boundaries, so these are matched with plain substring search.
const (
	thaiKaito    = "ไคโตะ"
	thaiMint     = "มิ้นต์"
	thaiMintAlt  = "มินต์" // alternative spelling without tone mark
	thaiTensei   = "เทนเซ"
	thaiTenseiAlt = "เท็นเซ" // alternative with mid-rising tone mark
)

// kanbanKeywords trigger personaKaito when no name is explicitly mentioned.
var kanbanKeywords = []string{
	// English
	"card", "board", "task", "todo", "to do", "to-do", "column", "kanban", "sprint", "backlog",
	// Thai transliterations / native
	"การ์ด", "บอร์ด", "แคนบัน", "คอลัมน์", "สปรินต์",
}

// financeKeywords trigger personaMint when no name is explicitly mentioned.
var financeKeywords = []string{
	// English
	"expense", "income", "spend", "spending", "budget", "salary",
	"baht", "฿", "category", "finance", "record", "cost", "payment",
	// Thai
	"บาท", "รายรับ", "รายจ่าย", "ค่าใช้จ่าย", "งบประมาณ", "เงินเดือน", "สรุปการเงิน",
}

// workoutKeywords trigger personaTensei when no name is explicitly mentioned.
// Covers workout, sleep, and nutrition domains — all handled by Tensei.
var workoutKeywords = []string{
	// English — workout
	"workout", "exercise", "training", "gym", "fitness", "strength",
	"cardio", "running", "sets", "reps", "weight training", "body weight",
	"mobility", "warmup", "warm-up", "cooldown", "cool-down", "stretching",
	"muscle", "preset", "session", "streak",
	// English — sleep & recovery
	"sleep", "sleeping", "bedtime", "wake up", "woke up", "nap",
	"sleep score", "sleep log", "insomnia", "recovery",
	// English — food & nutrition
	"food", "meal", "breakfast", "lunch", "dinner", "snack",
	"calorie", "calories", "protein", "carbs", "carbohydrate", "fat",
	"nutrition", "macro", "macros", "diet", "eating", "log food",
	// Thai — workout
	"ออกกำลังกาย", "ออกกำลัง", "ยิม", "ฟิตเนส", "ฝึกซ้อม", "วิ่ง", "น้ำหนัก", "กล้ามเนื้อ",
	// Thai — sleep & recovery
	"นอนหลับ", "การนอน", "นอน", "ตื่นนอน", "ก่อนนอน", "คะแนนการนอน", "พักผ่อน",
	// Thai — food & nutrition
	"อาหาร", "มื้ออาหาร", "อาหารเช้า", "อาหารกลางวัน", "อาหารเย็น", "ของว่าง",
	"แคลอรี่", "โปรตีน", "คาร์โบไฮเดรต", "ไขมัน", "โภชนาการ", "แมคโคร",
}

// detectPersona scans text for a persona signal.
// Name mention is checked first (both English and Thai); keyword fallback is applied
// when no name is found.
// Returns (persona, true) on a match, (_, false) when nothing matches.
func detectPersona(text string) (persona, bool) {
	if rKaito.MatchString(text) || strings.Contains(text, thaiKaito) {
		return personaKaito, true
	}
	if rMint.MatchString(text) || strings.Contains(text, thaiMint) || strings.Contains(text, thaiMintAlt) {
		return personaMint, true
	}
	if rTensei.MatchString(text) || strings.Contains(text, thaiTensei) || strings.Contains(text, thaiTenseiAlt) {
		return personaTensei, true
	}
	lower := strings.ToLower(text)
	for _, kw := range kanbanKeywords {
		if strings.Contains(lower, kw) {
			return personaKaito, true
		}
	}
	for _, kw := range financeKeywords {
		if strings.Contains(lower, kw) {
			return personaMint, true
		}
	}
	for _, kw := range workoutKeywords {
		if strings.Contains(lower, kw) {
			return personaTensei, true
		}
	}
	return personaAether, false
}

// resolvePersona determines which persona should handle the request.
//
// Order of precedence:
//  1. detectPersona on the current userMsg.
//  2. Walk history backward over user turns and use the most recently detected persona.
//  3. Default to personaAether.
func resolvePersona(history []Message, userMsg string) persona {
	if p, ok := detectPersona(userMsg); ok {
		return p
	}
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role != "user" {
			continue
		}
		if p, ok := detectPersona(history[i].Text); ok {
			return p
		}
	}
	return personaAether
}
