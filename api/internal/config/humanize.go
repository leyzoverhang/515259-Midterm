package config

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func Humanize(err error) string {
	return "invalid configuration:\n" + walk(err)
}

func walk(err error) string {
	if parsed, ok := err.(interface{ Unwrap() []error }); ok {
		var out []string
		for _, e := range parsed.Unwrap() {
			out = append(out, Humanize(e))
		}

		return strings.Join(out, "\n")
	}

	if verrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
		out := make([]string, 0, len(verrs))
		for _, fieldErr := range verrs {
			out = append(out, " - "+fieldErr.Namespace()+": ผิดกฎ "+fieldErr.Tag())
		}

		return strings.Join(out, "\n")
	}

	return " - " + err.Error()
}
