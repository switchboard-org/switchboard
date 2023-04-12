package parsecfg

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/switchboard-org/switchboard/providers"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/slices"
)

// switchboardBlockParser is responsible for parsing the global configuration
// that is passed in the switchboard block from user provided config.
type switchboardBlockParser struct {
	config     switchboardBlockStepConfig
	downloader providers.Downloader
}

// switchboardBlockStepConfig is a simple struct that allows us to parse the switchboard
// config block from the root of the config and leave the rest alone
type switchboardBlockStepConfig struct {
	Switchboard switchboardBlockContentsConfig `hcl:"switchboard,block"`
	Remain      hcl.Body                       `hcl:",remain"`
}

// switchboardBlockContentsConfig is the first data structure used to decode config data
// from the switchboard block, ignoring all but version and host for later processing
type switchboardBlockContentsConfig struct {
	//Version is an expression, so we can show diagnostics if necessary upon evaluation
	Version hcl.Expression `hcl:"version"`
	Remain  hcl.Body       `hcl:",remain"`
}

// requiredProviderBlocksStepConfig is a struct used for parsing required providers from the
// switchboard root block
type requiredProviderBlocksStepConfig struct {
	RequiredProviders []requiredProviderContentsConfig `hcl:"required_provider,block"`
}

// requireProviderBlockConfig
type requiredProviderContentsConfig struct {
	Name    string         `hcl:"name,label"`
	Source  string         `hcl:"source"`
	Version hcl.Expression `hcl:"version"`
	//there are no other fields. Remain gives us access the hcl.Range data for the block
	Remain hcl.Body `hcl:",remain"`
}

// requiredProviderData is a temporary data structure containing a fully parsed RequiredProviderBlock
// along with its config range, for debugging.
type requiredProviderData struct {
	block      RequiredProviderBlock
	blockRange hcl.Range
}

func (c *switchboardBlockParser) parse(currentVersion string, ctx *hcl.EvalContext, shouldVerifyDownloads bool) (*SwitchboardBlock, hcl.Diagnostics) {
	var diag hcl.Diagnostics
	versionStr, diag := c.parseVersion(c.config.Switchboard.Version, currentVersion, ctx)
	if diag.HasErrors() {
		return nil, diag
	}

	blocks, diag := c.parseRequiredBlocks(c.config.Switchboard.Remain, ctx)
	var requiredBlocks []RequiredProviderBlock
	if diag.HasErrors() {
		return nil, diag
	}

	if shouldVerifyDownloads {
		diag = c.verifyPresenceOfPackages(blocks)
		if diag.HasErrors() {
			return nil, diag
		}
	}

	for _, block := range blocks {
		requiredBlocks = append(requiredBlocks, block.block)
	}

	return &SwitchboardBlock{
			Version:           versionStr,
			RequiredProviders: requiredBlocks,
		},
		nil
}

// init is responsible for doing all related work at this part of the config when `switchboard init` is called
func (c *switchboardBlockParser) init(currentVersion string, ctx *hcl.EvalContext) (*SwitchboardBlock, hcl.Diagnostics) {
	var diag hcl.Diagnostics
	switchboardBlock, diag := c.parse(currentVersion, ctx, false)
	if diag.HasErrors() {
		return nil, diag
	}
	presentProviders, err := c.downloader.DownloadedProviders()
	if err != nil {
		reason := fmt.Sprintf("could not get list of downloaded providers. Reason: %s", err)
		diag = diag.Append(simpleDiagnostic(reason, reason, c.config.Switchboard.Remain.MissingItemRange()))
	}
	for _, provider := range switchboardBlock.RequiredProviders {
		providerPackage := providers.Package{
			Name:    providers.PackageName(provider.Source),
			Version: provider.Version,
		}
		if !slices.Contains(presentProviders, providerPackage) {
			err = c.downloader.DownloadProvider(provider.Source, provider.Version)
			if err != nil {
				reason := fmt.Sprintf("could not download required provider. Reason: %s", err)
				diag = diag.Append(simpleDiagnostic(reason, reason, c.config.Switchboard.Remain.MissingItemRange()))
			}
		}
	}
	return switchboardBlock, diag
}

// parseVersion is responsible for making sure the version condition set in the user config matches the current version
// of switchboard
func (c *switchboardBlockParser) parseVersion(expectedVersion hcl.Expression, currentVersion string, ctx *hcl.EvalContext) (string, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics
	if currentVersion == "development" {
		return currentVersion, diagnostics
	}
	currVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		diagnostics = diagnostics.Append(simpleDiagnostic(
			"Invalid version of CLI",
			fmt.Sprintf("Invalid version of CLI provided. If you are in development, set it to a valid semver value. Error: %s", err),
			c.config.Switchboard.Version.Range(),
		))
		return "", diagnostics
	}

	versionVal, _ := expectedVersion.Value(ctx)
	if versionVal.Type() != cty.String {
		diagnostics = diagnostics.Append(simpleDiagnostic(
			"Invalid type provided for version",
			fmt.Sprintf("Invalid version provided. Expected string. Got %s", versionVal.Type().GoString()),
			c.config.Switchboard.Version.Range(),
		))
	}
	expectedVersionConstraint, err := version.NewConstraint(versionVal.AsString())

	if err != nil {
		diagnostics = diagnostics.Append(simpleDiagnostic(
			"Invalid value provided for version",
			"Invalid value provided for version. Refer to 'github.com/hashicorp/go-version' documentation.",
			c.config.Switchboard.Version.Range(),
		))
	}

	if !expectedVersionConstraint.Check(currVersion) {
		diagnostics = diagnostics.Append(simpleDiagnostic(
			"Switchboard Version does not match expected version",
			fmt.Sprintf("Expected version to match constraint '%s'. Current version of Switchboard is %s", versionVal.AsString(), currentVersion),
			c.config.Switchboard.Version.Range(),
		))
	}
	return versionVal.AsString(), diagnostics
}

// parseRequiredBlocks is responsible for the simple parsing of the user config required_provider blocks.
func (c *switchboardBlockParser) parseRequiredBlocks(remain hcl.Body, ctx *hcl.EvalContext) ([]requiredProviderData, hcl.Diagnostics) {
	requiredPackageStepBlocks, diag := c.parseRequiredPackageBlocksStep(remain, ctx)
	if diag.HasErrors() {
		return nil, diag
	}

	var requiredPackageBlocks []requiredProviderData

	for _, provider := range requiredPackageStepBlocks.RequiredProviders {
		block, blockDiag := c.parseRequiredPackageBlockStep(provider, ctx)

		requiredPackageBlocks = append(requiredPackageBlocks, requiredProviderData{
			block:      block,
			blockRange: provider.Remain.MissingItemRange(),
		})
		diag = diag.Extend(blockDiag)
	}
	return requiredPackageBlocks, diag

}

// parseRequiredPackageBlocksStep parses the hcl.Body partial data into the appropriate data structure
// needed for the next step in the parsing process
func (c *switchboardBlockParser) parseRequiredPackageBlocksStep(providersBody hcl.Body, ctx *hcl.EvalContext) (requiredProviderBlocksStepConfig, hcl.Diagnostics) {
	var providerList requiredProviderBlocksStepConfig
	diag := gohcl.DecodeBody(providersBody, ctx, &providerList)
	return providerList, diag
}

// parseRequiredPackageBlockStep converts a temporary parsed struct into a RequiredProviderBlock,
// which is one of the final output type for required_provider blocks
func (c *switchboardBlockParser) parseRequiredPackageBlockStep(block requiredProviderContentsConfig, ctx *hcl.EvalContext) (RequiredProviderBlock, hcl.Diagnostics) {
	var packageVersion string
	diag := hcl.Diagnostics{}
	exprDiag := gohcl.DecodeExpression(block.Version, ctx, &packageVersion)
	if exprDiag.HasErrors() {
		return RequiredProviderBlock{}, diag.Extend(exprDiag)
	}
	return RequiredProviderBlock{
		Name:    block.Name,
		Source:  block.Source,
		Version: packageVersion,
	}, diag
}

// verifyPresenceOfPackages takes in a list of packages and checks whether they are present in the local
// package cache (usually in /.switchboard/packages/...)
func (c *switchboardBlockParser) verifyPresenceOfPackages(packages []requiredProviderData) hcl.Diagnostics {
	var diag hcl.Diagnostics
	downloadedProviders, err := c.downloader.DownloadedProviders()
	if err != nil {
		reason := fmt.Sprintf(" Packages could not be found. Run `switchboard init`")
		diag = diag.Append(simpleDiagnostic("Could not find downloaded packages", reason, c.config.Switchboard.Remain.MissingItemRange()))
	}

	for _, plugin := range packages {
		pack := providers.Package{
			Name:    providers.PackageName(plugin.block.Source),
			Version: plugin.block.Version,
		}
		if !slices.Contains(downloadedProviders, pack) {
			reason := fmt.Sprintf("plugin '%s', version '%s' is missing. Run `switchboard init` to download missing packages", plugin.block.Name, plugin.block.Version)
			diag = diag.Append(simpleDiagnostic("plugin package missing", reason, plugin.blockRange))
		}
	}
	return diag
}
