package providers

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/switchboard-org/switchboard/internal"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Package struct {
	Name    string
	Version string
}

type downloader struct {
	os            string
	arch          string
	packageFolder string
}

func defaultDownloader() downloader {
	return downloader{
		packageFolder: "./.switchboard/packages",
		os:            runtime.GOOS,
		arch:          runtime.GOARCH,
	}
}

// downloadedPackageList gets all packages that are currently
// downloaded from the cache
func (d *downloader) downloadedPackageList() ([]Package, error) {
	var packages []Package
	err := filepath.Walk(d.packageFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		pathSplit := strings.Split(path, "/")
		if !info.IsDir() && len(pathSplit) > 2 {
			packages = append(packages, Package{
				Name:    pathSplit[len(pathSplit)-3],
				Version: pathSplit[len(pathSplit)-2],
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return packages, nil

}

// packageIsDownloaded checks whether an existing package is downloaded
// to the package cache
func (d *downloader) packageIsDownloaded(location string, version string) (bool, error) {
	packageName := internal.PackageName(location)
	packageList, err := d.downloadedPackageList()
	if err != nil {
		return false, err
	}
	isDownloaded := false
	for _, pack := range packageList {
		if pack.Name == packageName && version == pack.Version {
			isDownloaded = true
		}
	}
	return isDownloaded, nil
}

// downloadPackage downloads a package Version from the location, which
// should be a public GitHub repository
func (d *downloader) downloadPackage(source string, version string) error {

	packageDistName, err := d.distName(source)
	if err != nil {
		return err
	}
	packagePath, err := d.packagePath(source, version)
	if err != nil {
		return err
	}

	var httpGetter = &getter.HttpGetter{
		ReadTimeout: 10 * time.Second,
	}
	getterClient := getter.Client{
		Src: fmt.Sprintf("https://%s/releases/download/v%s/%s", source, version, packageDistName),
		Dst: packagePath,
		Getters: map[string]getter.Getter{
			"http":  httpGetter,
			"https": httpGetter,
		},
		Mode: getter.ClientModeDir,
	}
	err = getterClient.Get()
	if err != nil {
		removeErr := os.RemoveAll(packagePath)
		if removeErr != nil {
			log.Printf("issue removing bad file. clear out your ./switchboard/packages directory and try again. Reason: %s\n", removeErr)
		}
	}
	return err
}

func (d *downloader) packagePath(source string, version string) (string, error) {
	return fmt.Sprintf("%s/%s/%s", d.packageFolder, source, version), nil
}

func (d *downloader) distName(location string) (string, error) {
	osName, err := d.systemOs()
	if err != nil {
		return "", err
	}
	processorName, err := d.processorArchitecture()
	zipFormat := "tar.gz"
	if osName == "windows" {
		zipFormat = "zip"
	}

	return fmt.Sprintf("%s_%s_%s.%s", internal.PackageName(location), internal.CapitalizeFirstLetterOfWord(osName), processorName, zipFormat), nil
}

func (d *downloader) systemOs() (string, error) {
	supportedOs := []string{"darwin", "linux", "windows"}
	if slices.Contains(supportedOs, d.os) {
		return d.os, nil
	}
	return "", errors.New(fmt.Sprintf("'%s' is not a supported OS", d.os))
}

func (d *downloader) processorArchitecture() (string, error) {
	switch d.arch {
	case "amd64":
		return "x86_64", nil
	case "386":
		return "i386", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported architecture: %s", d.arch)
	}
}
