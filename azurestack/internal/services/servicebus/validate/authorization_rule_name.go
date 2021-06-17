package validate

import (
	"regexp"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
)

func AuthorizationRuleName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile("^[a-zA-Z0-9][-._a-zA-Z0-9]{0,48}([a-zA-Z0-9])?$"),
		"The name can contain only letters, numbers, periods, hyphens and underscores. The name must start and end with a letter or number and be less the 50 characters long.",
	)
}
