package provider

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/net/context"
)

var _ validator.String = &stringUUIDValidator{}

type stringUUIDValidator struct{}

func (v stringUUIDValidator) Description(_ context.Context) string {
	return "the string parameter must be in a valid UUID"
}

func (v stringUUIDValidator) MarkdownDescription(_ context.Context) string {
	return "the string parameter must be in a valid UUID as defined here: [func Validate](https://pkg.go.dev/github.com/google/uuid#Validate)"
}

func (v stringUUIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if err := uuid.Validate(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"string attribute not a valid UUID",
			fmt.Sprintf("the provided string: %s is not a valid UUID", req.ConfigValue.ValueString()),
		)
	}
}

func stringUUIDValidate() stringUUIDValidator {
	return stringUUIDValidator{}
}
