// Package discover finds MPCs on the local network: hosts with SSH open that log in as root without a password and
// run the MPC app (Hakai firmware).
package discover

import (
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Found is an MPC on the network.
type Found struct {
	Host string `json:"host"`
	Name string `json:"name"` // the device's host name
}

// Scan checks every address of the computer's own /24 networks (in parallel) and returns the MPCs it finds.
func Scan() []Found {
	var hosts []string
	ifaces, _ := net.Interfaces()
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok || ipn.IP.To4() == nil || ipn.IP.IsLinkLocalUnicast() {
				continue
			}
			ip := ipn.IP.To4()
			for i := 1; i < 255; i++ {
				h := net.IPv4(ip[0], ip[1], ip[2], byte(i)).String()
				if h != ip.String() {
					hosts = append(hosts, h)
				}
			}
		}
	}
	var mu sync.Mutex
	var out []Found
	sem := make(chan struct{}, 64)
	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Add(1)
		sem <- struct{}{}
		go func(h string) {
			defer wg.Done()
			defer func() { <-sem }()
			if f, ok := probe(h); ok {
				mu.Lock()
				out = append(out, f)
				mu.Unlock()
			}
		}(h)
	}
	wg.Wait()
	sort.Slice(out, func(i, j int) bool { return out[i].Host < out[j].Host })
	return out
}

func probe(host string) (Found, bool) {
	c, err := net.DialTimeout("tcp", host+":22", 600*time.Millisecond)
	if err != nil {
		return Found{}, false
	}
	c.Close()
	cfg := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // only a probe: Link checks the key when it connects
		Timeout:         3 * time.Second,
	}
	cl, err := ssh.Dial("tcp", host+":22", cfg)
	if err != nil {
		return Found{}, false
	}
	defer cl.Close()
	s, err := cl.NewSession()
	if err != nil {
		return Found{}, false
	}
	defer s.Close()
	out, err := s.Output("test -x /usr/bin/MPC && hostname")
	if err != nil {
		return Found{}, false
	}
	return Found{Host: host, Name: strings.TrimSpace(string(out))}, true
}
