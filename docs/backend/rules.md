# Backend Rules

## Request validation

After binding an HTTP request with `c.BodyParser`, always validate using the shared
`pkg/validate` package (which wraps `github.com/go-playground/validator/v10`) — never
write manual validation helpers:

```go
func (h *Handler) createFoodLog(c *fiber.Ctx) error {
    var in health.CreateFoodLogInput
    if err := c.BodyParser(&in); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
    }
    if err := validate.Struct(in); err != nil {
        return err // *fiber.Error 400 with a descriptive field message
    }
    // ...
}
```

Constraints live on the input struct in `model.go` via `validate` struct tags:

```go
type CreateFoodLogInput struct {
    Name     string   `json:"name"      validate:"required"`
    MealType MealType `json:"meal_type" validate:"required,oneof=breakfast lunch dinner snack"`
    Calories *int     `json:"calories"  validate:"omitempty,min=0"`
}
```

- Use `oneof` for enum fields (values space-separated, matching the `type X string` constants).
- Use `required` for all mandatory fields; omit it for optional ones and handle the zero value intentionally.
- Use `omitempty` for optional pointer fields so validation is skipped when the field is nil/zero.
- The `pkg/validate` singleton is initialized once at startup — never construct `validator.New()` per request or per handler package.
- On failure, `validate.Struct` returns a `*fiber.Error` (400) naming the first failing field. The global `ErrorHandler` in `cmd/main.go` renders it automatically as `{"error": "..."}`.
- `github.com/go-playground/validator/v10` must be present in `go.mod`.

### Validation that tags cannot express

Keep logic in the handler for rules that struct tags cannot capture:

- Cross-field dependencies (e.g. "type is required when preset_id is nil").
- Semantic constraints with custom error messages (e.g. blocking a reserved enum value with an explanatory note).
- Post-bind transformations (e.g. defaulting an empty field to a sentinel value).
- Query/path param parsing — those are not body fields and must be checked manually.

## SQL queries — Jet only

**Never write raw SQL** (`db.QueryContext`, `db.ExecContext` with a plain string query) for any table that has generated Jet code in `gen/kinkando/public/`. Use the type-safe Jet builder exclusively.

```go
// Good — type-safe Jet query
stmt := postgres.SELECT(table.HealthWeightLogs.UserID).
    FROM(table.HealthWeightLogs).
    WHERE(table.HealthWeightLogs.LoggedAt.EQ(postgres.DateT(today)))
var rows []struct{ UserID uuid.UUID `alias:"health_weight_logs.user_id"` }
stmt.QueryContext(ctx, r.db, &rows)

// Bad — raw SQL string
rows, err := r.db.QueryContext(ctx,
    `SELECT user_id FROM health_weight_logs WHERE logged_at = $1`, today)
```

Jet covers the full query surface needed in this codebase:

- **INSERT … ON CONFLICT DO NOTHING**: `.ON_CONFLICT(cols...).DO_NOTHING()` — `ExecContext` returns `sql.Result` so you can check `RowsAffected()` for idempotency.
- **Subqueries in WHERE**: `col.NOT_IN(postgres.SELECT(...).FROM(...).WHERE(...))`.
- **LEFT JOIN + GROUP BY + custom aliases**: same pattern used in `quest/repository/repo.go::GetQuestStatus` and `CountIncompleteByUser`.
- **Cross-column JOIN predicates**: `table.A.ForeignKey.EQ(table.B.PrimaryKey)` — column-to-column comparison, not a literal value.

The only acceptable raw SQL is for tables that are outside the `gen/` codegen (e.g. `schema_migrations` managed by dbmate). When in doubt, check `gen/kinkando/public/table/` — if the file exists, use Jet.

## Date formatting

Use `time.RFC3339` instead of the literal `"2006-01-02T15:04:05Z07:00"` and
Use `time.DateOnly` instead of the literal `"2006-01-02"` in Go code:

```go
// Good
date, err := time.Parse(time.DateOnly, in.Date)
key := now.Format(time.DateOnly)

// Bad
date, err := time.Parse("2006-01-02", in.Date)
key := now.Format("2006-01-02")
```

**Exception:** struct tags are evaluated at compile time as raw string literals, so validator tags must keep the literal form:

```go
Birthdate *string `json:"birthdate" validate:"omitempty,datetime=2006-01-02"`
```
