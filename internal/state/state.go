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

type State interface {
	Load() error
	Save() error
	UpdateEntrie(dto.BinaryInfo)
	Has(string) bool
	Entries() map[string]dto.BinaryInfo
}

type LocalState struct {
	path     string
	file     *os.File
	Binaries map[string]dto.BinaryInfo
}

func NewState(path string) *LocalState {
	return &LocalState{
		path:     path,
		Binaries: make(map[string]dto.BinaryInfo),
	}
}

func (l *LocalState) Load() error {
	file, err := os.OpenFile(l.path, os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	l.file = file

	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		l.file.Close()
		return errors.New("another install/update command is currently running, try again later")
	}

	var binaries []dto.BinaryInfo
	err = json.NewDecoder(file).Decode(&binaries)
	if err != nil && !errors.Is(err, io.EOF) {
		l.file.Close()
		return err
	}

	for _, binaryInfo := range binaries {
		l.Binaries[binaryInfo.FullName] = binaryInfo
	}

	return nil
}

func (l *LocalState) UpdateEntrie(binaryInfo dto.BinaryInfo) {
	l.Binaries[binaryInfo.FullName] = binaryInfo
}

func (l *LocalState) Save() error {
	tmpPath := l.path + ".tmp"
	file, err := os.Create(filepath.Clean(tmpPath))
	if err != nil {
		return err
	}

	binaries := make([]dto.BinaryInfo, 0, len(l.Binaries))
	for _, binary := range l.Binaries {
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

	if err := os.Rename(tmpPath, l.path); err != nil {
		return err
	}

	if l.file != nil {
		_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
		return l.file.Close()
	}
	return nil
}

func (l *LocalState) Has(binaryName string) bool {
	_, ok := l.Binaries[binaryName]
	return ok
}

func (l *LocalState) Entries() map[string]dto.BinaryInfo {
	return l.Binaries
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
