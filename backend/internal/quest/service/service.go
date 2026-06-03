package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/quest"
)

type Repository interface {
	CreateQuest(ctx context.Context, userID uuid.UUID, in quest.CreateQuestInput) (*quest.Quest, error)
	ListQuests(ctx context.Context, userID uuid.UUID, questType string) ([]*quest.Quest, error)
	GetQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*quest.Quest, error)
	UpdateQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID, in quest.UpdateQuestInput) (*quest.Quest, error)
	DeleteQuest(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, userID uuid.UUID, active bool) (*quest.Quest, error)

	GetQuestStatus(ctx context.Context, userID uuid.UUID, questType quest.QuestType, today time.Time) ([]*quest.QuestStatus, error)
	TotalXP(ctx context.Context, userID uuid.UUID) (int, error)

	Increment(ctx context.Context, userID uuid.UUID, questID uuid.UUID, periodStart time.Time, source string) error
	Decrement(ctx context.Context, userID uuid.UUID, questID uuid.UUID, periodStart time.Time) error

	ProgressBySource(ctx context.Context, userID uuid.UUID, sourceType string, today, weekStart time.Time) error

	ListXPEvents(ctx context.Context, userID uuid.UUID, limit int) ([]*quest.XPEvent, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
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
	today := s.today()
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
	for _, d := range daily {
		if d.Completed {
			dailyDone++
		}
	}
	weeklyDone := 0
	for _, w := range weekly {
		if w.Completed {
			weeklyDone++
		}
	}

	return &quest.Overview{
		Date:        today.Format(time.DateOnly),
		WeekStart:   weekStart.Format(time.DateOnly),
		XP:          xpSummary(totalXP),
		Daily:       daily,
		Weekly:      weekly,
		DailyDone:   dailyDone,
		DailyTotal:  len(daily),
		WeeklyDone:  weeklyDone,
		WeeklyTotal: len(weekly),
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
	return s.repo.Increment(ctx, userID, questID, periodStart, source)
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
	return s.repo.ProgressBySource(ctx, userID, sourceType, s.today(), s.weekStart())
}

// periodFor returns the period start time and XP source label for the given quest type.
func (s *Service) periodFor(qType quest.QuestType) (time.Time, string) {
	if qType == quest.QuestTypeDaily {
		return s.today(), "daily"
	}
	return s.weekStart(), "weekly"
}

// ── History ───────────────────────────────────────────────────────────────────

func (s *Service) ListXPEvents(ctx context.Context, userID uuid.UUID, limit int) ([]*quest.XPEvent, error) {
	return s.repo.ListXPEvents(ctx, userID, limit)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// today returns midnight UTC for the current date in Asia/Bangkok timezone.
func (s *Service) today() time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// weekStart returns midnight UTC for the Monday that starts the current week in Asia/Bangkok.
func (s *Service) weekStart() time.Time {
	today := s.today()
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
