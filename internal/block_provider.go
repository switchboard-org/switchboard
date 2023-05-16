package internal

import "github.com/zclconf/go-cty/cty"

// ProviderBlock block lets a user configure various settings for a particular provider, such
// as auth contexts and other provider-specific settings. Refer to an individual provider plugin
// documentation for more details
type ProviderBlock struct {
	// BlockName will match the first label of the config block. As a convention, this name
	// should map to a required_provider. If not, 'required_provider' should be set to explicitly
	// tell switchboard what required provider it should map to.
	BlockName string
	// ProviderName is the actual provider this block maps to. Will be the same as BlockName
	// if 'required_provider' field was not set
	ProviderName string
	// InitPayload includes all the provider-specific data passed in the config. This data will
	// be passed to the plugin on initialization. The schema is also validated against what the provider
	// plugin expects to be here (which happens during the parsing step)
	InitPayload cty.Value
}
