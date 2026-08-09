package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() Validator {
	return Validator{validator: validator.New()}
}

func (v Validator) Validate(input any) error {
	err := v.validator.Struct(input)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return status.Error(codes.Internal, "internal validation error")
	}

	var msgs []string
	for _, e := range ve {
		field := e.Field()
		tag := e.Tag()
		param := e.Param()

		msg := ""
		switch {
		case field == "Email" && tag == "email":
			msg = "Invalid email format"
		case field == "Email" && tag == "required":
			msg = "Email is required"
		case field == "Email" && tag == "max":
			msg = fmt.Sprintf("Email must be at most %s characters", param)

		case field == "Password" && tag == "required":
			msg = "Password is required"
		case field == "Password" && tag == "min":
			msg = fmt.Sprintf("Password must be at least %s characters", param)
		case field == "Password" && tag == "max":
			msg = fmt.Sprintf("Password must be at most %s characters", param)

		case field == "Username" && tag == "alphanum":
			msg = "Username must contain only letters and numbers"
		case field == "Username" && tag == "min":
			msg = fmt.Sprintf("Username must be at least %s characters", param)
		case field == "Username" && tag == "max":
			msg = fmt.Sprintf("Username must be at most %s characters", param)

		case field == "AppID" && tag == "required":
			msg = "AppID is required"
		case field == "AppID" && (tag == "gt" || tag == "gte"):
			msg = "AppID must be a positive number"

		case field == "Roles" && tag == "min":
			msg = fmt.Sprintf("Roles list must contain at least %s items", param)
		case field == "Roles" && tag == "max":
			msg = fmt.Sprintf("Roles list must contain at most %s items", param)

		default:
			if param != "" {
				msg = fmt.Sprintf("%s field failed validation: %s (%s)", field, tag, param)
			} else {
				msg = fmt.Sprintf("%s field failed validation: %s", field, tag)
			}
		}
		msgs = append(msgs, msg)
	}

	return status.Errorf(codes.InvalidArgument, strings.Join(msgs, "; "))
}
