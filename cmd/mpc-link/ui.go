package main

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/systray"

	"github.com/FullPace/mpc-link/internal/config"
	"github.com/FullPace/mpc-link/internal/mpc"
)

// runUI runs the menu bar / tray app: status, settings page, connect / disconnect, credits, quit.
func runUI(cfg *config.Config) {
	app := newApp(cfg)
	if err := app.serve(); err != nil {
		fmt.Fprintln(os.Stderr, "settings page:", err)
		os.Exit(1)
	}
	systray.Run(func() { onReady(app) }, func() { app.link.Stop() })
}

func onReady(app *App) {
	systray.SetTemplateIcon(icon(false), icon(false))
	systray.SetTooltip("MPC Link")
	status := systray.AddMenuItem("Not connected", "")
	status.Disable()
	mount := systray.AddMenuItem("", "")
	mount.Disable()
	mount.Hide()
	systray.AddSeparator()
	settings := systray.AddMenuItem("Settings…", "Choose the MPC and the folders to share")
	toggle := systray.AddMenuItem("Disconnect", "")
	systray.AddSeparator()
	credits := systray.AddMenuItem("Credits…", "Made by Padbangers, licenses")
	quit := systray.AddMenuItem("Quit MPC Link", "Remove the folders from the MPC and quit")

	update := func(s mpc.Status) {
		st := app.state()
		switch {
		case !st.Enabled:
			status.SetTitle("Disconnected")
		case s.Connected:
			status.SetTitle("Connected to the MPC")
		case st.Host == "" || len(st.Folders) == 0:
			status.SetTitle("Not set up yet: open Settings")
		case s.Err != "":
			status.SetTitle("Not connected: " + shorten(s.Err, 48))
		default:
			status.SetTitle("Connecting…")
		}
		if s.Connected {
			mount.SetTitle(strings.TrimPrefix(s.Mount, "/media/"))
			mount.Show()
		} else {
			mount.Hide()
		}
		if st.Enabled {
			toggle.SetTitle("Disconnect")
		} else {
			toggle.SetTitle("Connect")
		}
		systray.SetTemplateIcon(icon(s.Connected), icon(s.Connected))
	}
	app.onState = update
	update(mpc.Status{})
	app.restart()
	if app.cfg.Host == "" || len(app.cfg.Folders) == 0 {
		openURL(app.url) // first start: straight to the settings
	}
	go func() {
		for {
			select {
			case <-settings.ClickedCh:
				openURL(app.url)
			case <-credits.ClickedCh:
				openURL(app.url + "&p=credits")
			case <-toggle.ClickedCh:
				app.setEnabled(!app.state().Enabled)
				update(app.link.Status())
			case <-quit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
