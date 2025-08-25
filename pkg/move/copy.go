package move

import (
	fsutils "github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/utils/fs"
	"github.com/go-logr/logr"
	"github.com/spf13/afero"
	"golang.org/x/sys/unix"
)

const (
	noPermissionsMask = 0000
)

type CopyFunc func(log logr.Logger, fs afero.Afero, from, to string) error

var _ CopyFunc = SimpleCopy

func SimpleCopy(log logr.Logger, fs afero.Afero, from, to string) error {
	log.Info("starting to copy (simple)", "from", from, "to", to)

	oldUmask := unix.Umask(noPermissionsMask)
	defer unix.Umask(oldUmask)

	err := fsutils.CopyFolder(log, fs, from, to)
	if err != nil {
		log.Error(err, "error moving folder")

		return err
	}

	log.Info("successfully copied (simple)", "from", from, "to", to)

	return nil
}
