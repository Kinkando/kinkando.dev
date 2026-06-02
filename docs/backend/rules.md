# Backend Rules

## Request validation

After binding an HTTP request with `c.BodyParser`, always validate using `github.com/go-playground/validator/v10` — never write manual validation helpers:

```go
var validate = validator.New()

func (h *Handler) createFoodLog(c *fiber.Ctx) error {
    var in health.CreateFoodLogInput
    if err := c.BodyParser(&in); err != nil {
        return fiber.ErrBadRequest
    }
    if err := validate.Struct(in); err != nil {
        return fiber.ErrBadRequest
    }
    // ...
}
```

Constraints live on the input struct in `model.go` via `validate` struct tags:

```go
type CreateFoodLogInput struct {
    Name     string   `json:"name"     validate:"required"`
    MealType MealType `json:"meal_type" validate:"required,oneof=breakfast lunch dinner snack"`
    Calories int      `json:"calories"  validate:"required,min=0"`
}
```

- Use `oneof` for enum fields (values space-separated, matching the `type X string` constants).
- Use `required` for all mandatory fields; omit it for optional ones and handle the zero value intentionally.
- The validator instance should be a package-level or handler-struct-level singleton — not constructed per request.
- `github.com/go-playground/validator/v10` must be present in `go.mod`. Add it with `go get github.com/go-playground/validator/v10` if missing.
