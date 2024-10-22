package provider

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/net/context"
)

var _ validator.String = &stringUUIDValidator{}

type stringUUIDValidator struct{}

func (v stringUUIDValidator) Description(_ context.Context) string {
	return "the string parameter must be in a valid UUID"
}

func (v stringUUIDValidator) MarkdownDescription(_ context.Context) string {
	return "the string parameter must be a valid UUID as defined here: [func Validate](https://pkg.go.dev/github.com/google/uuid#Validate)"
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

// AtLeastSumOfValidator ensures that the length is at least the sum of the provided minimums.
type AtLeastSumOfValidator struct {
	minLowercaseAttr string
	minUppercaseAttr string
	minNumberAttr    string
	minSpecialAttr   string
}

func (v AtLeastSumOfValidator) Description(_ context.Context) string {
	return "the length attibute must be greater than or equal to the sum of min_lowercase, min_uppercase, min_number, and min_special."
}

func (v AtLeastSumOfValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v AtLeastSumOfValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Extract the values of the min_* attributes
	var minLowercase, minUppercase, minNumber, minSpecial types.Int64

	diags := req.Config.GetAttribute(ctx, path.Root(v.minLowercaseAttr), &minLowercase)
	resp.Diagnostics.Append(diags...)

	diags = req.Config.GetAttribute(ctx, path.Root(v.minUppercaseAttr), &minUppercase)
	resp.Diagnostics.Append(diags...)

	diags = req.Config.GetAttribute(ctx, path.Root(v.minNumberAttr), &minNumber)
	resp.Diagnostics.Append(diags...)

	diags = req.Config.GetAttribute(ctx, path.Root(v.minSpecialAttr), &minSpecial)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Sum the minimums
	sumMin := minLowercase.ValueInt64() + minUppercase.ValueInt64() + minNumber.ValueInt64() + minSpecial.ValueInt64()

	// Validate that the length is at least the sum of the minimums
	if req.ConfigValue.ValueInt64() < sumMin {
		resp.Diagnostics.AddError(
			"Invalid length",
			fmt.Sprintf("The length (%d) must be at least the sum of min_lowercase (%d), min_uppercase (%d), min_number (%d), and min_special (%d), which is %d.",
				req.ConfigValue.ValueInt64(), minLowercase, minUppercase, minNumber, minSpecial, sumMin),
		)
	}
}

// AtLeastSumOf returns a custom validator for the length attribute.
func AtLeastSumOf(minLowercaseAttr, minUppercaseAttr, minNumberAttr, minSpecialAttr string) validator.Int64 {
	return AtLeastSumOfValidator{
		minLowercaseAttr: minLowercaseAttr,
		minUppercaseAttr: minUppercaseAttr,
		minNumberAttr:    minNumberAttr,
		minSpecialAttr:   minSpecialAttr,
	}
}

// ConditionalMinValidator validates a `min_*` field only if the corresponding boolean field is true.
type ConditionalMinValidator struct {
	booleanAttr string // The attribute name for the corresponding boolean field (e.g., "lowercase", "numbers", etc.)
	minAttrName string // The attribute name for the `min_*` field being validated
}

// ValidateInt64 checks whether the `min_*` field should be validated based on the boolean flag.
func (v ConditionalMinValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	//if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
	//	return
	//}

	// Retrieve the value of the corresponding boolean attribute
	var booleanEnabled types.Bool
	diags := req.Config.GetAttribute(ctx, path.Root(v.booleanAttr), &booleanEnabled)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Perform validation only if the boolean flag is true
	if booleanEnabled.ValueBool() && (req.ConfigValue.ValueInt64() < 1 || req.ConfigValue.ValueInt64() > 9) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Invalid %s", v.minAttrName),
			fmt.Sprintf("%s must be between 1 and 9 when %s is enabled.", v.minAttrName, v.booleanAttr),
		)
	}
}

// Description provides a description for the validator.
func (v ConditionalMinValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Validates %s if %s is true.", v.minAttrName, v.booleanAttr)
}

// MarkdownDescription provides a markdown description for the validator.
func (v ConditionalMinValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ConditionalMin returns a new instance of the custom validator for any `min_*` field.
func ConditionalMin(booleanAttr, minAttrName string) validator.Int64 {
	return ConditionalMinValidator{
		booleanAttr: booleanAttr,
		minAttrName: minAttrName,
	}
}
