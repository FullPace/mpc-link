// Package share serves the shared computer folders to the MPC over SFTP.
//
// The MPC mounts the whole share with sshfs; its root lists one directory per shared folder (by name), and every
// path below maps into that folder on the computer. Nothing outside the shared folders can be reached.
package share

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/sftp"

	"github.com/FullPace/mpc-link/internal/config"
)

type fs struct {
	folders map[string]string // name -> computer path
}

// Serve runs an SFTP server for the folders over rw (an SSH session's stdin/stdout) until it closes.
func Serve(rw io.ReadWriteCloser, folders []config.Folder) error {
	f := &fs{folders: map[string]string{}}
	for _, d := range folders {
		f.folders[d.Name] = d.Path
	}
	h := sftp.Handlers{FileGet: f, FilePut: f, FileCmd: f, FileList: f}
	srv := sftp.NewRequestServer(rw, h)
	err := srv.Serve()
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// resolve maps an SFTP path to a computer path. root is true for "/" (the virtual folder list).
func (f *fs) resolve(p string) (local string, root bool, err error) {
	clean := path.Clean("/" + p)
	if clean == "/" {
		return "", true, nil
	}
	parts := strings.SplitN(strings.TrimPrefix(clean, "/"), "/", 2)
	base, ok := f.folders[parts[0]]
	if !ok {
		return "", false, os.ErrNotExist
	}
	if len(parts) == 1 {
		return base, false, nil
	}
	return filepath.Join(base, filepath.FromSlash(parts[1])), false, nil // Clean above removed any ".."
}

func (f *fs) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	p, root, err := f.resolve(r.Filepath)
	if err != nil {
		return nil, err
	}
	if root {
		return nil, os.ErrPermission
	}
	return os.Open(p)
}

func (f *fs) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	p, root, err := f.resolve(r.Filepath)
	if err != nil {
		return nil, err
	}
	if root {
		return nil, os.ErrPermission
	}
	flags := os.O_RDWR | os.O_CREATE
	if r.Pflags().Trunc {
		flags |= os.O_TRUNC
	}
	return os.OpenFile(p, flags, 0o644)
}

func (f *fs) Filecmd(r *sftp.Request) error {
	p, root, err := f.resolve(r.Filepath)
	if err != nil {
		return err
	}
	if root {
		return os.ErrPermission
	}
	switch r.Method {
	case "Setstat":
		a := r.Attributes()
		if r.AttrFlags().Size {
			if err := os.Truncate(p, int64(a.Size)); err != nil {
				return err
			}
		}
		if r.AttrFlags().Acmodtime {
			return os.Chtimes(p, time.Unix(int64(a.Atime), 0), time.Unix(int64(a.Mtime), 0))
		}
		return nil
	case "Rename", "PosixRename":
		t, troot, err := f.resolve(r.Target)
		if err != nil {
			return err
		}
		if troot {
			return os.ErrPermission
		}
		return os.Rename(p, t)
	case "Rmdir":
		return os.Remove(p)
	case "Remove":
		return os.Remove(p)
	case "Mkdir":
		return os.Mkdir(p, 0o755)
	case "Symlink", "Link":
		return os.ErrPermission
	}
	return sftp.ErrSSHFxOpUnsupported
}

type listAt []os.FileInfo

func (l listAt) ListAt(out []os.FileInfo, off int64) (int, error) {
	if off >= int64(len(l)) {
		return 0, io.EOF
	}
	n := copy(out, l[off:])
	if n < len(out) {
		return n, io.EOF
	}
	return n, nil
}

// named renames a FileInfo (the folder entries of the virtual root carry their share name).
type named struct {
	os.FileInfo
	name string
}

func (n named) Name() string { return n.name }

type rootInfo struct{}

func (rootInfo) Name() string       { return "/" }
func (rootInfo) Size() int64        { return 0 }
func (rootInfo) Mode() os.FileMode  { return os.ModeDir | 0o755 }
func (rootInfo) ModTime() time.Time { return time.Now() }
func (rootInfo) IsDir() bool        { return true }
func (rootInfo) Sys() interface{}   { return nil }

func (f *fs) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	p, root, err := f.resolve(r.Filepath)
	if err != nil {
		return nil, err
	}
	switch r.Method {
	case "List":
		if root {
			names := make([]string, 0, len(f.folders))
			for n := range f.folders {
				names = append(names, n)
			}
			sort.Strings(names)
			var out listAt
			for _, n := range names {
				if fi, err := os.Stat(f.folders[n]); err == nil && fi.IsDir() {
					out = append(out, named{fi, n})
				}
			}
			return out, nil
		}
		entries, err := os.ReadDir(p)
		if err != nil {
			return nil, err
		}
		var out listAt
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "._") || e.Name() == ".DS_Store" {
				continue // macOS metadata files: noise in the MPC's browser
			}
			if fi, err := e.Info(); err == nil {
				out = append(out, fi)
			}
		}
		return out, nil
	case "Stat", "Lstat":
		if root {
			return listAt{rootInfo{}}, nil
		}
		fi, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if base := path.Base(path.Clean("/" + r.Filepath)); base != fi.Name() {
			return listAt{named{fi, base}}, nil
		}
		return listAt{fi}, nil
	}
	return nil, sftp.ErrSSHFxOpUnsupported
}
