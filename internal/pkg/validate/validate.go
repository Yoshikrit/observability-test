package validate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var instance = validator.New()

// Struct validates s against its `validate` struct tags and returns a single
// human-readable error listing every failed field, or nil.
func Struct(s any) error {
	if err := instance.Struct(s); err != nil {
		return errors.New(formatError(err))
	}
	return nil
}

// FiberValidator adapts Struct to fiber.StructValidator, so c.Bind().Body(&req)
// parses and validates in one call - wire it in via fiber.Config{StructValidator: ...}.
type FiberValidator struct{}

func (FiberValidator) Validate(out any) error {
	return Struct(out)
}

func formatError(err error) string {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		msgs := make([]string, 0, len(validationErrs))
		for _, fe := range validationErrs {
			msgs = append(msgs, fmt.Sprintf("%s failed on '%s'", fe.Field(), fe.Tag()))
		}
		return strings.Join(msgs, "; ")
	}
	return "invalid request body"
}
