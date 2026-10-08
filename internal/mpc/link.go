// Package mpc connects to an MPC over SSH and mounts the computer's shared folders on it.
//
// The MPC (Hakai firmware) runs OpenSSH with root login and an empty password. Padbangers Link copies a static
// sshfs to /data/hacks/link/ and runs it in passive mode: sshfs speaks SFTP over the SSH session's stdin/stdout,
// and this side answers with share.Serve. So the computer needs no SSH server and no open port. The mount is a
// folder on a drive the MPC already shows (MPC SSD/<computer>, else MPC Documents/<computer>): the MPC's browser
// only lists drives its own disk daemon mounted, but shows any folder on them.
package mpc

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/FullPace/padbangers-link/internal/config"
	"github.com/FullPace/padbangers-link/internal/share"
)

const (
	remoteDir   = "/data/hacks/link"
	remoteSSHFS = remoteDir + "/sshfs"
)

// mountBases, in order of preference: a folder on the SSD, else in the internal MPC Documents.
var mountBases = []string{"/media/MPC SSD", "/media/az01-internal-sd/MPC Documents"}

// SSHFS is the armv7 sshfs binary (static, see device/), uploaded to the MPC when missing or different.
var SSHFS []byte

// Quote makes s a single shell word for the MPC's BusyBox sh.
func Quote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// Dial opens an SSH connection to the MPC as root. The host key is remembered on first use (*hostKey empty) and
// checked afterwards, like ssh's known_hosts. Only the Ed25519 key is used: the MPC makes a new ECDSA key on every
// boot (in the /etc overlay), while its Ed25519 key comes with the firmware image and stays. A remembered key of
// another type (from version 0.1.0, which took the ECDSA key) is replaced once.
func Dial(host string, hostKey *string) (*ssh.Client, error) {
	cfg := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(""),
			ssh.KeyboardInteractive(func(string, string, []string, []bool) ([]string, error) { return nil, nil }),
		},
		HostKeyAlgorithms: []string{ssh.KeyAlgoED25519},
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			got := base64.StdEncoding.EncodeToString(key.Marshal())
			if *hostKey == "" || !strings.HasPrefix(*hostKey, "AAAAC3NzaC1lZDI1NTE5") { // empty, or not an Ed25519 key
				*hostKey = got
				return nil
			}
			if got != *hostKey {
				return errors.New("the MPC's SSH host key changed (a different device at this address?)")
			}
			return nil
		},
		Timeout: 8 * time.Second,
	}
	if !strings.Contains(host, ":") {
		host += ":22"
	}
	return ssh.Dial("tcp", host, cfg)
}

// Run runs a shell command on the MPC and returns its output.
func Run(c *ssh.Client, cmd string) (string, error) {
	s, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer s.Close()
	var out bytes.Buffer
	s.Stdout = &out
	s.Stderr = &out
	err = s.Run(cmd)
	return out.String(), err
}

// EnsureSSHFS uploads sshfs when the MPC has none or a different one.
func EnsureSSHFS(c *ssh.Client) error {
	sum := md5.Sum(SSHFS)
	want := hex.EncodeToString(sum[:])
	if out, _ := Run(c, "md5sum "+remoteSSHFS+" 2>/dev/null"); strings.HasPrefix(out, want) {
		return nil
	}
	if _, err := Run(c, "mkdir -p "+remoteDir); err != nil {
		return err
	}
	sc, err := sftp.NewClient(c)
	if err != nil {
		return err
	}
	defer sc.Close()
	tmp := remoteSSHFS + ".new"
	f, err := sc.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := f.Write(SSHFS); err != nil {
		f.Close()
		return err
	}
	f.Close()
	_, err = Run(c, fmt.Sprintf("chmod 755 %s && mv %s %s", tmp, tmp, remoteSSHFS))
	return err
}

// MountPoint picks where the computer's folder goes on the MPC.
func MountPoint(c *ssh.Client, name string) (string, error) {
	for _, b := range mountBases {
		if _, err := Run(c, "test -d "+Quote(b)); err == nil {
			return b + "/" + name, nil
		}
	}
	return "", errors.New("no MPC SSD or MPC Documents folder found on the MPC")
}

// Unmount removes a mount and its (empty) folder; harmless when nothing is mounted.
func Unmount(c *ssh.Client, mnt string) {
	Run(c, fmt.Sprintf("umount -l %s 2>/dev/null; rmdir %s 2>/dev/null", Quote(mnt), Quote(mnt)))
}

// Mount mounts the shared folders at mnt and serves them until the session ends or stop is closed.
func Mount(c *ssh.Client, mnt string, folders []config.Folder, stop <-chan struct{}) error {
	Unmount(c, mnt) // a mount left from a lost connection
	s, err := c.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()
	in, err := s.StdinPipe()
	if err != nil {
		return err
	}
	out, err := s.StdoutPipe()
	if err != nil {
		return err
	}
	var errBuf bytes.Buffer
	s.Stderr = &errBuf
	cmd := fmt.Sprintf("mkdir -p %s && exec %s -f -o passive,allow_other ':/' %s",
		Quote(mnt), remoteSSHFS, Quote(mnt))
	if err := s.Start(cmd); err != nil {
		return err
	}
	done := make(chan error, 2)
	go func() { done <- share.Serve(&pipe{out, in}, folders) }()
	go func() { done <- s.Wait() }()
	select {
	case err = <-done:
	case <-stop:
		err = nil
	}
	s.Close()
	Unmount(c, mnt)
	if err != nil && errBuf.Len() > 0 {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(errBuf.String()))
	}
	return err
}

// pipe joins an SSH session's stdout (requests from sshfs) and stdin (our replies).
type pipe struct {
	r io.Reader
	w io.WriteCloser
}

func (p *pipe) Read(b []byte) (int, error)  { return p.r.Read(b) }
func (p *pipe) Write(b []byte) (int, error) { return p.w.Write(b) }
func (p *pipe) Close() error                { return p.w.Close() }

// Status of a Link, for the menu bar.
type Status struct {
	Connected bool   `json:"connected"`
	Mount     string `json:"mount"`
	Err       string `json:"err"`
}

// Link keeps the folders mounted on the MPC: it connects, mounts, and reconnects after a lost connection.
type Link struct {
	mu      sync.Mutex
	status  Status
	stop    chan struct{}
	stopped chan struct{}
	OnState func(Status) // called on every change (from the link's goroutine)
}

func (l *Link) Status() Status {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.status
}

func (l *Link) set(s Status) {
	l.mu.Lock()
	l.status = s
	l.mu.Unlock()
	if l.OnState != nil {
		l.OnState(s)
	}
}

// Start runs the link with cfg until Stop. save is called when the host key was learned.
func (l *Link) Start(cfg *config.Config, name string, save func()) {
	l.Stop()
	l.stop = make(chan struct{})
	l.stopped = make(chan struct{})
	stop, stopped := l.stop, l.stopped
	go func() {
		defer close(stopped)
		wait := 2 * time.Second
		for {
			err := l.once(cfg, name, save, stop)
			select {
			case <-stop:
				l.set(Status{})
				return
			default:
			}
			msg := "disconnected"
			if err != nil {
				msg = err.Error()
			}
			l.set(Status{Err: msg})
			select {
			case <-stop:
				l.set(Status{})
				return
			case <-time.After(wait):
			}
			if wait < 30*time.Second {
				wait *= 2
			}
		}
	}()
}

func (l *Link) once(cfg *config.Config, name string, save func(), stop <-chan struct{}) error {
	if cfg.Host == "" {
		return errors.New("no MPC address set")
	}
	known := cfg.HostKey
	c, err := Dial(cfg.Host, &cfg.HostKey)
	if err != nil {
		return err
	}
	defer c.Close()
	if known != cfg.HostKey {
		save()
	}
	if err := EnsureSSHFS(c); err != nil {
		return fmt.Errorf("installing sshfs on the MPC: %v", err)
	}
	mnt, err := MountPoint(c, name)
	if err != nil {
		return err
	}
	l.set(Status{Connected: true, Mount: mnt})
	// keepalives notice a vanished MPC (switched off, out of Wi-Fi) within ~30 s
	alive := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		misses := 0
		for {
			select {
			case <-alive:
				return
			case <-t.C:
				if _, _, err := c.SendRequest("keepalive@openssh.com", true, nil); err != nil {
					if misses++; misses >= 3 {
						c.Close()
						return
					}
				} else {
					misses = 0
				}
			}
		}
	}()
	defer close(alive)
	return Mount(c, mnt, cfg.Folders, stop)
}

// Stop disconnects and removes the mount from the MPC.
func (l *Link) Stop() {
	if l.stop == nil {
		return
	}
	close(l.stop)
	<-l.stopped
	l.stop = nil
}
