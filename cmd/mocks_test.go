package cmd

import (
	"errors"

	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
)

const (
	TestBinaryName     = "foo"
	TestBinaryFullName = "foo/foo"
	TestBinaryVersion  = "0.0.0"
	TestResolvedURL    = "https://fake.url/foo.tar.gz"

	DummyStateErrorMessage     = "some error in state"
	DummyResolverErrorMessage  = "some error in resolver"
	DummyInstallerErrorMessage = "some error in installer"
)

type DummyState struct {
	loadCount int
	saveCount int
	onError   bool

	binaries map[string]dto.BinaryInfo
}

func (s *DummyState) Load() error {
	s.loadCount++
	if s.onError {
		return errors.New(DummyStateErrorMessage)
	}
	return nil
}

func (s *DummyState) Save() error {
	s.saveCount++
	if s.onError {
		return errors.New(DummyStateErrorMessage)
	}
	return nil
}

func (s *DummyState) UpdateEntrie(binaryInfo dto.BinaryInfo) {
	s.binaries[binaryInfo.FullName] = binaryInfo
}

func (s *DummyState) Has(binaryName string) bool {
	_, ok := s.binaries[binaryName]
	return ok
}

func (s *DummyState) Entries() map[string]dto.BinaryInfo {
	return s.binaries
}

type DummyResolver struct {
	resolveCount int
	onError      bool
}

func (r *DummyResolver) Resolve(*dto.BinaryInfo) (string, error) {
	r.resolveCount++
	if r.onError {
		return "", errors.New(DummyResolverErrorMessage)
	}
	return TestResolvedURL, nil
}

type DummyInstaller struct {
	installCount int
	onError      bool
}

func (i *DummyInstaller) Install(binaryInfo *dto.BinaryInfo, url string) error {
	i.installCount++
	if i.onError {
		return errors.New(DummyInstallerErrorMessage)
	}
	binaryInfo.InstalledVersion = TestBinaryVersion
	return nil
}
