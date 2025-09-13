package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
)

const StateFileName = "state.json"

type State struct {
	path     string
	file     *os.File
	Binaries map[string]dto.BinaryInfo
}

func NewState(path string) *State {
	return &State{
		path:     path,
		Binaries: make(map[string]dto.BinaryInfo),
	}
}

func (s *State) Load() error {
	file, err := os.OpenFile(s.path, os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	s.file = file

	if err := syscall.Flock(int(s.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		s.file.Close()
		return errors.New("another install/update command is currently running, try again later")
	}

	var binaries []dto.BinaryInfo
	err = json.NewDecoder(file).Decode(&binaries)
	if err != nil && !errors.Is(err, io.EOF) {
		s.file.Close()
		return err
	}

	for _, binaryInfo := range binaries {
		s.Binaries[binaryInfo.FullName] = binaryInfo
	}

	return nil
}

func (s *State) UpdateEntrie(binaryInfo dto.BinaryInfo) {
	s.Binaries[binaryInfo.FullName] = binaryInfo
}

func (s *State) Save() error {
	tmpPath := s.path + ".tmp"
	file, err := os.Create(filepath.Clean(tmpPath))
	if err != nil {
		return err
	}

	binaries := make([]dto.BinaryInfo, 0, len(s.Binaries))
	for _, binary := range s.Binaries {
		binaries = append(binaries, binary)
	}

	encErr := json.NewEncoder(file).Encode(binaries)
	syncErr := file.Sync()
	closeErr := file.Close()
	if encErr != nil || syncErr != nil || closeErr != nil {
		os.Remove(tmpPath)
		if encErr != nil {
			return encErr
		}
		if syncErr != nil {
			return syncErr
		}
		return closeErr
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return err
	}

	if s.file != nil {
		_ = syscall.Flock(int(s.file.Fd()), syscall.LOCK_UN)
		return s.file.Close()
	}
	return nil
}

func (s *State) Has(binaryName string) bool {
	_, ok := s.Binaries[binaryName]
	return ok
}

func StateDirectory() string {
	var mkdirErr error
	cfgDir, err := os.UserConfigDir()
	stateDir := filepath.Join(cfgDir, "azabox")
	if err == nil {
		mkdirErr = os.MkdirAll(filepath.Clean(stateDir), 0o755)
	}
	if err != nil || mkdirErr != nil {
		fmt.Fprintln(os.Stderr, "Cannot determine user config dir, aborting")
		os.Exit(1)
	}
	return stateDir
}
