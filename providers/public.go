package providers

import "strings"

//func FetchProvider(name string, version string) error {
//
//}

type Downloader interface {
	// DownloadedProviders returns a list of currently downloaded providers
	DownloadedProviders() ([]Package, error)
	// DownloadProvider fetches the provider from the source and downloads it into the
	// provider cache, as defined by the downloader implementation
	DownloadProvider(source string, version string) error
	// ProviderPath is a helpful method for telling us where a provider is located
	// so that it can be called as a plugin
	ProviderPath(source string, version string) string
}

type DefaultDownloader struct {
	downloader downloader
}

func NewDefaultDownloader() Downloader {
	return &DefaultDownloader{
		downloader: defaultDownloader(),
	}
}

func (d *DefaultDownloader) DownloadedProviders() ([]Package, error) {
	return d.downloader.downloadedPackageList()
}

func (d *DefaultDownloader) DownloadProvider(source string, version string) error {
	return d.downloader.downloadPackage(source, version)
}

func (d *DefaultDownloader) ProviderPath(source string, version string) string {
	return d.downloader.packagePath(source, version) + "/" + PackageName(source)
}

func PackageName(source string) string {
	packageNameParts := strings.Split(source, "/")
	return packageNameParts[len(packageNameParts)-1]
}
