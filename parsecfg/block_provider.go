package parsecfg

import "github.com/hashicorp/hcl/v2"

// ProviderBlock block lets a user configure various settings for a particular provider, such
// as auth contexts and other provider-specific settings. A ProviderBlock block will be mapped to a
// RequiredProviderBlock block by matching "name" block labels.
type ProviderBlock struct {
	Name           string                       `hcl:"name,label"`
	Authorizations []ProviderAuthorizationBlock `hcl:"authorization,block"`
	//providers can have their own schemas for provider blocks. They will fall into the Remain field
	Remain hcl.Body `hcl:",remain"`
}

type ProviderAuthorizationBlock struct {
	Name string `hcl:"name,label"`
}
