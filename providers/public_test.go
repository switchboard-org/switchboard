package providers

type MockDownloader struct {
	packages []Package
}

func NewTestDownloader(packages []Package) Downloader {
	return &MockDownloader{
		packages: packages,
	}
}

func (d *MockDownloader) DownloadedProviders() ([]Package, error) {
	return d.packages, nil
}

func (d *MockDownloader) DownloadProvider(_ string, _ string) error {
	return nil
}

func (d *MockDownloader) ProviderPath(_ string, _ string) string {
	return ".mock-packages"
}
