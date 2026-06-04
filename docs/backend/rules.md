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

## Date formatting

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
