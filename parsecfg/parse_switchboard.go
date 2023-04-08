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
	config     switchboardStepConfig
	downloader providers.Downloader
}

// switchboardStepConfig is a simple struct that allows us to parse the switchboard
// config block from the root of the config and leave the rest alone
type switchboardStepConfig struct {
	Switchboard switchboardStepBlockConfig `hcl:"switchboard,block"`
	Remain      hcl.Body                   `hcl:",remain"`
}

// switchboardStepBlockConfig is the first data structure used to decode config data
// from the switchboard block, ignoring all but version and host for later processing
type switchboardStepBlockConfig struct {
	//Version is an expression, so we can show diagnostics if necessary upon evaluation
	Version hcl.Expression `hcl:"version"`
	Host    HostBlock      `hcl:"host,block"`
	Remain  hcl.Body       `hcl:",remain"`
}

// switchboardStepProviderBlocksConfig is isolated by itself to access details for
// diagnostics for error handling
type switchboardStepProviderBlocksConfig struct {
	RequiredProviders []requiredProviderBlockConfig `hcl:"required_provider,block"`
}

type requiredProviderBlockConfig struct {
	Name    string         `hcl:"name,label"`
	Source  string         `hcl:"source"`
	Version hcl.Expression `hcl:"version"`
	//this is used to give us access to the hcl range
	Remain hcl.Body `hcl:",remain"`
}
type requiredProviderBlock struct {
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
	for _, block := range blocks {
		requiredBlocks = append(requiredBlocks, block.block)
	}
	if diag.HasErrors() {
		return nil, diag
	}

	if shouldVerifyDownloads {
		diag = c.verifyPresenceOfPackage(blocks)
		if diag.HasErrors() {
			return nil, diag
		}
	}

	return &SwitchboardBlock{
			Version:           versionStr,
			Host:              c.config.Switchboard.Host,
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

func (c *switchboardBlockParser) parseRequiredBlocks(remain hcl.Body, ctx *hcl.EvalContext) ([]requiredProviderBlock, hcl.Diagnostics) {
	requiredPackageStepBlocks, diag := c.parseRequiredPackageBlocksStep(remain, ctx)
	if diag.HasErrors() {
		return nil, diag
	}

	var requiredPackageBlocks []requiredProviderBlock

	for _, provider := range requiredPackageStepBlocks.RequiredProviders {
		block, blockDiag := c.parseRequiredPackageBlockStep(provider, ctx)

		requiredPackageBlocks = append(requiredPackageBlocks, requiredProviderBlock{
			block:      block,
			blockRange: provider.Remain.MissingItemRange(),
		})
		diag = diag.Extend(blockDiag)
	}
	return requiredPackageBlocks, diag

}

func (c *switchboardBlockParser) parseRequiredPackageBlocksStep(providersBody hcl.Body, ctx *hcl.EvalContext) (switchboardStepProviderBlocksConfig, hcl.Diagnostics) {
	var providerList switchboardStepProviderBlocksConfig
	diag := gohcl.DecodeBody(providersBody, ctx, &providerList)
	return providerList, diag
}

func (c *switchboardBlockParser) parseRequiredPackageBlockStep(block requiredProviderBlockConfig, ctx *hcl.EvalContext) (RequiredProviderBlock, hcl.Diagnostics) {
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

func (c *switchboardBlockParser) verifyPresenceOfPackage(packages []requiredProviderBlock) hcl.Diagnostics {
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
