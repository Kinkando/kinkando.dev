package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/quest"
	"github.com/kinkando/personal-dashboard/pkg/event"
	"github.com/kinkando/personal-dashboard/pkg/helper"
)

// EventPublisher is the narrow interface quest depends on for emitting domain
// events. *event.Bus satisfies it; quest never imports subscriber packages.
type EventPublisher interface {
	Publish(ctx context.Context, e event.Event)
}

type Repository interface {
	CreateQuest(ctx context.Context, userID uuid.UUID, in quest.CreateQuestInput) (*quest.Quest, error)
	ListQuests(ctx context.Context, userID uuid.UUID, questType string) ([]*quest.Quest, error)
	GetQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*quest.Quest, error)
	UpdateQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID, in quest.UpdateQuestInput) (*quest.Quest, error)
	DeleteQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, userID uuid.UUID, active bool) (*quest.Quest, error)

	GetQuestStatus(ctx context.Context, userID uuid.UUID, questType quest.QuestType, today time.Time) ([]*quest.QuestStatus, error)
	TotalXP(ctx context.Context, userID uuid.UUID) (int, error)
	ListDailyResults(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]quest.PeriodResultRow, error)

	Increment(ctx context.Context, userID uuid.UUID, questID uuid.UUID, periodStart time.Time, source string) error
	Decrement(ctx context.Context, userID uuid.UUID, questID uuid.UUID, periodStart time.Time) error

	ProgressBySource(ctx context.Context, userID uuid.UUID, sourceType string, today, weekStart time.Time) error

	ListXPEvents(ctx context.Context, userID uuid.UUID, limit int) ([]*quest.XPEvent, error)
}

type Service struct {
	repo   Repository
	events EventPublisher   // nil-safe; set via New
	now    func() time.Time // injectable clock; defaults to helper.Today
}

func New(repo Repository, events EventPublisher) *Service {
	return &Service{repo: repo, events: events, now: helper.Today}
}

// publishCompleted emits QuestCompleted so cross-module subscribers (e.g.
// achievements) can re-evaluate level/XP/quest-count milestones. nil-safe.
func (s *Service) publishCompleted(ctx context.Context, userID uuid.UUID) {
	if s.events != nil {
		s.events.Publish(ctx, event.Event{Type: event.QuestCompleted, UserID: userID})
	}
}

// ── Quest CRUD ────────────────────────────────────────────────────────────────

func (s *Service) CreateQuest(ctx context.Context, userID uuid.UUID, in quest.CreateQuestInput) (*quest.Quest, error) {
	// Default source_type to manual when omitted.
	if in.SourceType == "" {
		in.SourceType = quest.SourceTypeManual
	}
	return s.repo.CreateQuest(ctx, userID, in)
}

func (s *Service) ListQuests(ctx context.Context, userID uuid.UUID, questType string) ([]*quest.Quest, error) {
	return s.repo.ListQuests(ctx, userID, questType)
}

func (s *Service) UpdateQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID, in quest.UpdateQuestInput) (*quest.Quest, error) {
	return s.repo.UpdateQuest(ctx, id, userID, in)
}

func (s *Service) DeleteQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteQuest(ctx, id, userID)
}

func (s *Service) SetActive(ctx context.Context, id uuid.UUID, userID uuid.UUID, active bool) (*quest.Quest, error) {
	return s.repo.SetActive(ctx, id, userID, active)
}

// ── Overview ──────────────────────────────────────────────────────────────────

func (s *Service) GetOverview(ctx context.Context, userID uuid.UUID) (*quest.Overview, error) {
	today := s.now()
	weekStart := s.weekStart()

	daily, err := s.repo.GetQuestStatus(ctx, userID, quest.QuestTypeDaily, today)
	if err != nil {
		return nil, err
	}
	weekly, err := s.repo.GetQuestStatus(ctx, userID, quest.QuestTypeWeekly, weekStart)
	if err != nil {
		return nil, err
	}
	totalXP, err := s.repo.TotalXP(ctx, userID)
	if err != nil {
		return nil, err
	}

	if daily == nil {
		daily = []*quest.QuestStatus{}
	}
	if weekly == nil {
		weekly = []*quest.QuestStatus{}
	}

	dailyDone := 0
	var dailyQuests []*quest.QuestStatus
	for index, d := range daily {
		if !d.IsActive {
			continue
		}
		dailyQuests = append(dailyQuests, daily[index])
		if d.Completed {
			dailyDone++
		}
	}
	weeklyDone := 0
	var weeklyQuests []*quest.QuestStatus
	for index, w := range weekly {
		if !w.IsActive {
			continue
		}
		weeklyQuests = append(weeklyQuests, weekly[index])
		if w.Completed {
			weeklyDone++
		}
	}

	return &quest.Overview{
		Date:          today.Format(time.DateOnly),
		WeekStart:     weekStart.Format(time.DateOnly),
		XP:            xpSummary(totalXP),
		Daily:         dailyQuests,
		Weekly:        weeklyQuests,
		DailyDone:     dailyDone,
		DailyTotal:    len(dailyQuests),
		WeeklyDone:    weeklyDone,
		WeeklyTotal:   len(weeklyQuests),
		DailyBonusXP:  quest.DailyBonusXP,
		WeeklyBonusXP: quest.WeeklyBonusXP,
	}, nil
}

// ── Streaks ───────────────────────────────────────────────────────────────────

// GetStreaks builds the daily-quest consistency heatmap and streak counters for
// the trailing 365 days. The quest_period_results snapshot only holds finalized
// (past) days, so today's live status is merged on top via GetQuestStatus. A
// "perfect day" is one where every active daily quest reached its target; a day
// with no active daily quests is a gap that breaks a streak.
func (s *Service) GetStreaks(ctx context.Context, userID uuid.UUID) (*quest.StreakSummary, error) {
	today := s.now()
	from := today.AddDate(0, 0, -364)

	rows, err := s.repo.ListDailyResults(ctx, userID, from, today)
	if err != nil {
		return nil, err
	}

	// Aggregate snapshot rows by day → {total, completed}.
	type agg struct{ total, completed int }
	byDay := map[string]*agg{}
	for _, r := range rows {
		k := r.PeriodStart.Format(time.DateOnly)
		a := byDay[k]
		if a == nil {
			a = &agg{}
			byDay[k] = a
		}
		a.total++
		if r.Completed {
			a.completed++
		}
	}

	// Merge today's live status (the snapshot table has no today row).
	liveDaily, err := s.repo.GetQuestStatus(ctx, userID, quest.QuestTypeDaily, today)
	if err != nil {
		return nil, err
	}
	if len(liveDaily) > 0 {
		a := &agg{total: len(liveDaily)}
		for _, q := range liveDaily {
			if q.Completed {
				a.completed++
			}
		}
		byDay[today.Format(time.DateOnly)] = a // override or insert
	}

	perfect := func(a *agg) bool { return a != nil && a.total > 0 && a.completed >= a.total }

	// Build cells (only days with active quests) and count perfect days.
	var days []quest.HeatmapDay
	perfectDays := 0
	for d := from; !d.After(today); d = d.AddDate(0, 0, 1) {
		if a := byDay[d.Format(time.DateOnly)]; a != nil && a.total > 0 {
			days = append(days, quest.HeatmapDay{Date: d.Format(time.DateOnly), Total: a.total, Completed: a.completed})
			if perfect(a) {
				perfectDays++
			}
		}
	}

	// Current streak: walk back from today; if today isn't yet perfect, start at
	// yesterday so the in-progress day doesn't break the chain.
	current := 0
	start := today
	if !perfect(byDay[today.Format(time.DateOnly)]) {
		start = today.AddDate(0, 0, -1)
	}
	for d := start; !d.Before(from); d = d.AddDate(0, 0, -1) {
		if perfect(byDay[d.Format(time.DateOnly)]) {
			current++
		} else {
			break
		}
	}

	// Longest streak across the window (a gap / zero-total day breaks the run).
	longest, run := 0, 0
	for d := from; !d.After(today); d = d.AddDate(0, 0, 1) {
		if perfect(byDay[d.Format(time.DateOnly)]) {
			run++
			if run > longest {
				longest = run
			}
		} else {
			run = 0
		}
	}

	return &quest.StreakSummary{
		Days:          days,
		CurrentStreak: current,
		LongestStreak: longest,
		PerfectDays:   perfectDays,
	}, nil
}

// ── Actions ───────────────────────────────────────────────────────────────────

// IncrementQuest adds one completion for the quest's current period.
// Only manual quests may be incremented by the user; auto quests are driven by HandleSourceEvent.
func (s *Service) IncrementQuest(ctx context.Context, userID uuid.UUID, questID uuid.UUID) error {
	q, err := s.repo.GetQuest(ctx, questID, userID)
	if err != nil {
		return err
	}
	if q.SourceType != quest.SourceTypeManual {
		return questError("this quest is auto-managed and cannot be manually incremented")
	}
	periodStart, source := s.periodFor(q.Type)
	if err := s.repo.Increment(ctx, userID, questID, periodStart, source); err != nil {
		return err
	}
	s.publishCompleted(ctx, userID)
	return nil
}

// DecrementQuest removes the most recent completion for the quest's current period.
// Only manual quests may be decremented by the user.
func (s *Service) DecrementQuest(ctx context.Context, userID uuid.UUID, questID uuid.UUID) error {
	q, err := s.repo.GetQuest(ctx, questID, userID)
	if err != nil {
		return err
	}
	if q.SourceType != quest.SourceTypeManual {
		return questError("this quest is auto-managed and cannot be manually decremented")
	}
	periodStart, _ := s.periodFor(q.Type)
	return s.repo.Decrement(ctx, userID, questID, periodStart)
}

// HandleSourceEvent advances all active quests linked to the given sourceType
// for the user. Called by the event bus — never by the user directly.
func (s *Service) HandleSourceEvent(ctx context.Context, userID uuid.UUID, sourceType string) error {
	if err := s.repo.ProgressBySource(ctx, userID, sourceType, s.now(), s.weekStart()); err != nil {
		return err
	}
	s.publishCompleted(ctx, userID)
	return nil
}

// periodFor returns the period start time and XP source label for the given quest type.
func (s *Service) periodFor(qType quest.QuestType) (time.Time, string) {
	if qType == quest.QuestTypeDaily {
		return s.now(), "daily"
	}
	return s.weekStart(), "weekly"
}

// ── History ───────────────────────────────────────────────────────────────────

func (s *Service) ListXPEvents(ctx context.Context, userID uuid.UUID, limit int) ([]*quest.XPEvent, error) {
	return s.repo.ListXPEvents(ctx, userID, limit)
}

// weekStart returns midnight UTC for the Monday that starts the current week in Asia/Bangkok.
func (s *Service) weekStart() time.Time {
	today := s.now()
	weekday := today.Weekday()
	daysFromMonday := int(weekday-time.Monday+7) % 7
	return today.AddDate(0, 0, -daysFromMonday)
}

// xpSummary derives level and progress from raw total XP.
func xpSummary(totalXP int) quest.XPSummary {
	const xpPerLevel = 100
	level := totalXP/xpPerLevel + 1
	xpInto := totalXP % xpPerLevel
	return quest.XPSummary{
		TotalXP:     totalXP,
		Level:       level,
		XPIntoLevel: xpInto,
		XPForLevel:  xpPerLevel,
		XPToNext:    xpPerLevel - xpInto,
	}
}

// questError wraps a plain string into an error for consistent return style.
func questError(msg string) error {
	return &questErr{msg}
}

type questErr struct{ msg string }

func (e *questErr) Error() string { return e.msg }
