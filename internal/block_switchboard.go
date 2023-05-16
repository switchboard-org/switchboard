package internal

// SwitchboardBlock contains the primary global configuration elements of all workflows,
// including the required providers, log settings, retry settings, and more.
type SwitchboardBlock struct {
	Version string
	//Host              HostBlock
	RequiredProviders []RequiredProviderBlock
}

// RequiredProviderBlock tells us where a provider should be pulled from, and which version it
// should use.
type RequiredProviderBlock struct {
	Name    string
	Source  string
	Version string
}

//// HostBlock tells us where the workflow runner is hosted and the api key to trigger deployments
//type HostBlock struct {
//	Address string `hcl:"address"`
//	Key     string `hcl:"key"`
//}
