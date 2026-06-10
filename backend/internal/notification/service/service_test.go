package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/internal/fcm"
	"github.com/kinkando/personal-dashboard/internal/line"
	"github.com/kinkando/personal-dashboard/internal/notification"
	"github.com/kinkando/personal-dashboard/internal/user"
	"go.uber.org/zap"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

func ptr[T any](v T) *T { return &v }

type fakeRepo struct {
	settings    *notification.Settings
	tokens      []string
	deletedToks []string
}

func (f *fakeRepo) GetSettings(_ context.Context, _ uuid.UUID) (*notification.Settings, error) {
	return f.settings, nil
}
func (f *fakeRepo) UpsertSettings(_ context.Context, _ uuid.UUID, in notification.UpsertSettingsInput) (*notification.Settings, error) {
	return &notification.Settings{}, nil
}
func (f *fakeRepo) AddToken(_ context.Context, _ uuid.UUID, _ string) error    { return nil }
func (f *fakeRepo) DeleteToken(_ context.Context, tok string) error {
	f.deletedToks = append(f.deletedToks, tok)
	return nil
}
func (f *fakeRepo) ListTokens(_ context.Context, _ uuid.UUID) ([]string, error) {
	return f.tokens, nil
}
func (f *fakeRepo) HasToken(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return false, nil
}

type fakeLine struct {
	calls []string // "to:text"
}

func (f *fakeLine) Push(_ context.Context, to string, msgs []line.ReplyMessage) error {
	f.calls = append(f.calls, to)
	return nil
}

type fakeDiscord struct {
	calls []string // content strings
}

func (f *fakeDiscord) PostWebhook(_ context.Context, _, content string) error {
	f.calls = append(f.calls, content)
	return nil
}

type fakeFCM struct {
	invalidTokens map[string]bool // tokens that should return ErrTokenInvalid
	calls         []string
}

func (f *fakeFCM) Send(_ context.Context, tok, _, _ string) error {
	f.calls = append(f.calls, tok)
	if f.invalidTokens[tok] {
		return fcm.ErrTokenInvalid
	}
	return nil
}

type fakeUsers struct {
	u *user.User
}

func (f *fakeUsers) GetByID(_ context.Context, _ uuid.UUID) (*user.User, error) {
	return f.u, nil
}

func newSvc(repo *fakeRepo, ln LinePusher, disc DiscordSender, fcmSender FCMSender, users *fakeUsers) *Service {
	return &Service{
		repo:    repo,
		line:    ln,
		discord: disc,
		fcm:     fcmSender,
		users:   users,
		log:     zap.NewNop(),
	}
}

// ── Settings nil / missing ────────────────────────────────────────────────────

func TestNotify_SettingsNil_NoDelivery(t *testing.T) {
	// repo returns nil settings (user has never configured notifications).
	repo := &fakeRepo{settings: nil}
	ln := &fakeLine{}
	svc := newSvc(repo, ln, &fakeDiscord{}, &fakeFCM{}, &fakeUsers{})

	res := svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "T", Body: "B"})

	if res.Attempted != 0 {
		t.Errorf("Attempted = %d, want 0 (nil settings)", res.Attempted)
	}
	if len(ln.calls) != 0 {
		t.Error("LINE should not be called when settings is nil")
	}
}

// ── LINE channel ──────────────────────────────────────────────────────────────

func TestNotify_LINE_DeliversWhenEnabled(t *testing.T) {
	lineID := "Uabc123"
	settings := &notification.Settings{LineEnabled: true}
	repo := &fakeRepo{settings: settings}
	ln := &fakeLine{}
	users := &fakeUsers{u: &user.User{LineID: &lineID}}
	svc := newSvc(repo, ln, &fakeDiscord{}, &fakeFCM{}, users)

	res := svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "Hi", Body: "Hello"})

	if res.Attempted != 1 {
		t.Errorf("Attempted = %d, want 1", res.Attempted)
	}
	if res.Delivered != 1 {
		t.Errorf("Delivered = %d, want 1", res.Delivered)
	}
	if len(ln.calls) != 1 || ln.calls[0] != lineID {
		t.Errorf("LINE Push called with %v, want [%s]", ln.calls, lineID)
	}
}

func TestNotify_LINE_SkippedWhenSenderNil(t *testing.T) {
	lineID := "Uabc123"
	settings := &notification.Settings{LineEnabled: true}
	repo := &fakeRepo{settings: settings}
	users := &fakeUsers{u: &user.User{LineID: &lineID}}
	svc := newSvc(repo, nil, &fakeDiscord{}, &fakeFCM{}, users) // nil LINE sender

	res := svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "T", Body: "B"})

	if res.Attempted != 0 {
		t.Errorf("Attempted = %d, want 0 (nil LINE sender)", res.Attempted)
	}
}

// ── Discord channel ───────────────────────────────────────────────────────────

func TestNotify_Discord_DeliversWhenEnabled(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/test"
	settings := &notification.Settings{DiscordEnabled: true, DiscordWebhookURL: &webhookURL}
	repo := &fakeRepo{settings: settings}
	disc := &fakeDiscord{}
	svc := newSvc(repo, &fakeLine{}, disc, &fakeFCM{}, &fakeUsers{})

	res := svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "News", Body: "An update"})

	if res.Delivered != 1 {
		t.Errorf("Delivered = %d, want 1", res.Delivered)
	}
	if len(disc.calls) != 1 {
		t.Fatalf("Discord calls = %d, want 1", len(disc.calls))
	}
	if !strings.Contains(disc.calls[0], "News") || !strings.Contains(disc.calls[0], "An update") {
		t.Errorf("Discord content = %q, should contain title and body", disc.calls[0])
	}
}

func TestNotify_Discord_SkippedWhenWebhookURLNil(t *testing.T) {
	settings := &notification.Settings{DiscordEnabled: true, DiscordWebhookURL: nil}
	repo := &fakeRepo{settings: settings}
	disc := &fakeDiscord{}
	svc := newSvc(repo, &fakeLine{}, disc, &fakeFCM{}, &fakeUsers{})

	svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "T", Body: "B"})

	if len(disc.calls) != 0 {
		t.Error("Discord should be skipped when webhook URL is nil")
	}
}

// ── FCM / Web Push channel ────────────────────────────────────────────────────

func TestNotify_FCM_DeliversToAllTokens(t *testing.T) {
	settings := &notification.Settings{WebPushEnabled: true}
	tokens := []string{"tok1", "tok2"}
	repo := &fakeRepo{settings: settings, tokens: tokens}
	fcmSender := &fakeFCM{}
	svc := newSvc(repo, &fakeLine{}, &fakeDiscord{}, fcmSender, &fakeUsers{})

	res := svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "Push", Body: "msg"})

	if res.Attempted != 2 {
		t.Errorf("Attempted = %d, want 2 (one per token)", res.Attempted)
	}
	if res.Delivered != 2 {
		t.Errorf("Delivered = %d, want 2", res.Delivered)
	}
}

func TestNotify_FCM_PrunesInvalidToken(t *testing.T) {
	settings := &notification.Settings{WebPushEnabled: true}
	tokens := []string{"valid-tok", "stale-tok"}
	repo := &fakeRepo{settings: settings, tokens: tokens}
	fcmSender := &fakeFCM{invalidTokens: map[string]bool{"stale-tok": true}}
	svc := newSvc(repo, &fakeLine{}, &fakeDiscord{}, fcmSender, &fakeUsers{})

	res := svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "T", Body: "B"})

	// stale-tok should have been pruned.
	if len(repo.deletedToks) != 1 || repo.deletedToks[0] != "stale-tok" {
		t.Errorf("deletedToks = %v, want [stale-tok]", repo.deletedToks)
	}
	// Only valid-tok succeeded.
	if res.Delivered != 1 {
		t.Errorf("Delivered = %d, want 1 (only valid token)", res.Delivered)
	}
	// The result must contain an error for the pruned token.
	if len(res.Errors) == 0 {
		t.Error("expected at least one error entry for the invalid token")
	}
}

// ── Message formatting ────────────────────────────────────────────────────────

func TestNotify_Discord_ContentFormatsTitle(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/x"
	settings := &notification.Settings{DiscordEnabled: true, DiscordWebhookURL: &webhookURL}
	repo := &fakeRepo{settings: settings}
	disc := &fakeDiscord{}
	svc := newSvc(repo, &fakeLine{}, disc, &fakeFCM{}, &fakeUsers{})

	svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "MyTitle", Body: "MyBody"})

	if len(disc.calls) == 0 {
		t.Fatal("no Discord call")
	}
	// Title must be bold (**title**) in Discord markdown.
	if !strings.Contains(disc.calls[0], fmt.Sprintf("**%s**", "MyTitle")) {
		t.Errorf("Discord content %q should bold the title", disc.calls[0])
	}
}

// ── Error paths ───────────────────────────────────────────────────────────────

func TestNotify_Discord_ErrorRecorded(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/x"
	settings := &notification.Settings{DiscordEnabled: true, DiscordWebhookURL: &webhookURL}
	repo := &fakeRepo{settings: settings}
	errDisc := &errDiscord{}
	svc := newSvc(repo, &fakeLine{}, errDisc, &fakeFCM{}, &fakeUsers{})

	res := svc.Notify(context.Background(), uuid.New(), notification.Message{Title: "T", Body: "B"})

	if res.Delivered != 0 {
		t.Errorf("Delivered = %d, want 0 (Discord error)", res.Delivered)
	}
	if len(res.Errors) == 0 {
		t.Error("expected error entry for Discord failure")
	}
}

// errDiscord always returns an error.
type errDiscord struct{}

func (e *errDiscord) PostWebhook(_ context.Context, _, _ string) error {
	return errors.New("webhook rejected")
}
