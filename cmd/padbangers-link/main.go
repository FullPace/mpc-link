// Padbangers Link: the computer's folders in the MPC's browser, over the network.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/FullPace/padbangers-link/device"
	"github.com/FullPace/padbangers-link/internal/config"
	"github.com/FullPace/padbangers-link/internal/mpc"
)

func computerName(cfg *config.Config) string {
	if cfg.ComputerName != "" {
		return config.SafeName(cfg.ComputerName)
	}
	return config.SafeName(hostName())
}

func main() {
	host := flag.String("host", "", "the MPC's address (saved)")
	var shares multi
	flag.Var(&shares, "share", "share a folder: [Name=]/path (repeatable, saved)")
	headless := flag.Bool("headless", false, "no menu bar icon: run in the terminal until Ctrl-C")
	flag.Parse()
	mpc.SSHFS = device.SSHFS

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "settings:", err)
		os.Exit(1)
	}
	if *host != "" {
		cfg.Host = *host
	}
	for _, s := range shares {
		name, p, ok := strings.Cut(s, "=")
		if !ok {
			p, name = s, filepath.Base(s)
		}
		cfg.Folders = append(cfg.Folders, config.Folder{Name: config.SafeName(name), Path: p})
	}
	if *host != "" || len(shares) > 0 {
		cfg.Save()
	}

	if !*headless {
		runUI(cfg)
		return
	}
	link := &mpc.Link{OnState: func(s mpc.Status) {
		switch {
		case s.Connected:
			fmt.Println("connected:", s.Mount)
		case s.Err != "":
			fmt.Println("not connected:", s.Err)
		}
	}}
	link.Start(cfg, computerName(cfg), func() { cfg.Save() })
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Println("disconnecting")
	link.Stop()
}

type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(s string) error { *m = append(*m, s); return nil }
