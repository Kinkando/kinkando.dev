package achievement

import "time"

// Metric identifies the underlying number a badge is measured against.
type Metric string

const (
	MetricLevel           Metric = "level"
	MetricTotalXP         Metric = "total_xp"
	MetricCurrentStreak   Metric = "current_streak"
	MetricLongestStreak   Metric = "longest_streak"
	MetricPerfectDays     Metric = "perfect_days"
	MetricWorkouts        Metric = "workouts"
	MetricWeightLogs      Metric = "weight_logs"
	MetricSleepLogs       Metric = "sleep_logs"
	MetricMedicineIntakes Metric = "medicine_intakes"
	MetricQuestsCompleted Metric = "quests_completed"
)

// Def is a single badge definition. The catalog lives in code; only unlock state
// is persisted (user_achievements). Code is a stable key — never reuse or rename.
type Def struct {
	Code        string
	Title       string
	Description string
	Icon        string // emoji
	Category    string
	Metric      Metric
	Threshold   int
}

// Achievement is a badge plus the requesting user's progress toward it.
type Achievement struct {
	Code        string     `json:"code"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Category    string     `json:"category"`
	Unlocked    bool       `json:"unlocked"`
	UnlockedAt  *time.Time `json:"unlocked_at"`
	Progress    int        `json:"progress"`
	Target      int        `json:"target"`
}

// Summary is the GET /achievements payload.
type Summary struct {
	Achievements  []Achievement `json:"achievements"`
	UnlockedCount int           `json:"unlocked_count"`
	Total         int           `json:"total"`
	NewlyUnlocked []string      `json:"newly_unlocked"` // codes unlocked during this evaluation
}

// Catalog is the fixed set of badges. Ordered by category then threshold so the
// API returns them ready to render. Codes are permanent identifiers.
var Catalog = []Def{
	// ── Rank (Adventure Rank / level) ──────────────────────────────────────────
	{Code: "level_5", Title: "Apprentice", Description: "Reach Level 5", Icon: "🥉", Category: "Rank", Metric: MetricLevel, Threshold: 5},
	{Code: "level_10", Title: "Adept", Description: "Reach Level 10", Icon: "🥈", Category: "Rank", Metric: MetricLevel, Threshold: 10},
	{Code: "level_25", Title: "Veteran", Description: "Reach Level 25", Icon: "🥇", Category: "Rank", Metric: MetricLevel, Threshold: 25},
	{Code: "level_50", Title: "Master", Description: "Reach Level 50", Icon: "👑", Category: "Rank", Metric: MetricLevel, Threshold: 50},

	// ── Consistency (streaks & perfect days) ───────────────────────────────────
	{Code: "streak_7", Title: "Consistent", Description: "A 7-day perfect streak", Icon: "🔥", Category: "Consistency", Metric: MetricLongestStreak, Threshold: 7},
	{Code: "streak_30", Title: "Unstoppable", Description: "A 30-day perfect streak", Icon: "⚡", Category: "Consistency", Metric: MetricLongestStreak, Threshold: 30},
	{Code: "streak_100", Title: "Centurion", Description: "A 100-day perfect streak", Icon: "💯", Category: "Consistency", Metric: MetricLongestStreak, Threshold: 100},
	{Code: "perfect_10", Title: "Dedicated", Description: "10 perfect days total", Icon: "✨", Category: "Consistency", Metric: MetricPerfectDays, Threshold: 10},
	{Code: "perfect_50", Title: "Devoted", Description: "50 perfect days total", Icon: "🌟", Category: "Consistency", Metric: MetricPerfectDays, Threshold: 50},
	{Code: "perfect_100", Title: "Relentless", Description: "100 perfect days total", Icon: "💫", Category: "Consistency", Metric: MetricPerfectDays, Threshold: 100},

	// ── Fitness (workouts) ──────────────────────────────────────────────────────
	{Code: "workouts_1", Title: "First Sweat", Description: "Complete your first workout", Icon: "💪", Category: "Fitness", Metric: MetricWorkouts, Threshold: 1},
	{Code: "workouts_10", Title: "Getting Strong", Description: "Complete 10 workouts", Icon: "🏋️", Category: "Fitness", Metric: MetricWorkouts, Threshold: 10},
	{Code: "workouts_50", Title: "Gym Rat", Description: "Complete 50 workouts", Icon: "🏆", Category: "Fitness", Metric: MetricWorkouts, Threshold: 50},
	{Code: "workouts_100", Title: "Iron Will", Description: "Complete 100 workouts", Icon: "🦾", Category: "Fitness", Metric: MetricWorkouts, Threshold: 100},

	// ── Health (sleep, weight, medicine) ────────────────────────────────────────
	{Code: "sleep_7", Title: "Well Rested", Description: "Log 7 nights of sleep", Icon: "😴", Category: "Health", Metric: MetricSleepLogs, Threshold: 7},
	{Code: "sleep_30", Title: "Sleep Tracker", Description: "Log 30 nights of sleep", Icon: "🌙", Category: "Health", Metric: MetricSleepLogs, Threshold: 30},
	{Code: "sleep_100", Title: "Dream Keeper", Description: "Log 100 nights of sleep", Icon: "🛌", Category: "Health", Metric: MetricSleepLogs, Threshold: 100},
	{Code: "weight_1", Title: "On the Scale", Description: "Log your first weigh-in", Icon: "⚖️", Category: "Health", Metric: MetricWeightLogs, Threshold: 1},
	{Code: "weight_30", Title: "Tracking Progress", Description: "Log 30 weigh-ins", Icon: "📉", Category: "Health", Metric: MetricWeightLogs, Threshold: 30},
	{Code: "weight_100", Title: "Data Driven", Description: "Log 100 weigh-ins", Icon: "📊", Category: "Health", Metric: MetricWeightLogs, Threshold: 100},
	{Code: "meds_50", Title: "On Schedule", Description: "Record 50 medicine intakes", Icon: "💊", Category: "Health", Metric: MetricMedicineIntakes, Threshold: 50},
	{Code: "meds_200", Title: "Never Miss", Description: "Record 200 medicine intakes", Icon: "🩺", Category: "Health", Metric: MetricMedicineIntakes, Threshold: 200},

	// ── Quests ──────────────────────────────────────────────────────────────────
	{Code: "quests_10", Title: "Questing", Description: "Complete 10 quests", Icon: "📜", Category: "Quests", Metric: MetricQuestsCompleted, Threshold: 10},
	{Code: "quests_100", Title: "Quest Hunter", Description: "Complete 100 quests", Icon: "🗺️", Category: "Quests", Metric: MetricQuestsCompleted, Threshold: 100},
	{Code: "quests_500", Title: "Legend", Description: "Complete 500 quests", Icon: "⭐", Category: "Quests", Metric: MetricQuestsCompleted, Threshold: 500},
}

var byCode = func() map[string]Def {
	m := make(map[string]Def, len(Catalog))
	for _, d := range Catalog {
		m[d.Code] = d
	}
	return m
}()

// FindDef returns the catalog definition for a badge code.
func FindDef(code string) (Def, bool) {
	d, ok := byCode[code]
	return d, ok
}
