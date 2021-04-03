package main

import (
	"flag"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

const (
	RES240      = "240p"
	RES360      = "360p"
	RES480      = "480p"
	RES720      = "720p"
	CONFIG_FILE = ".meeting-media"
	WM          = "WM"
	MM          = "MM"
)

func main() {
	config := NewConfig()
	a := app.New()

	config.DebugMode = flag.Bool("d", false, "fake downloading; print debug info")
	flag.Parse()
	if *config.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("RUNNING IN DEBUG MODE")
	}

	progressBar := widget.NewProgressBar()
	config.Progress = &progress{0, "", progressBar}
	pbFormatter := func() string { return config.Progress.Title }
	config.Progress.ProgressBar.TextFormatter = pbFormatter

	settingsTab := container.NewTabItem("", config.settingsGUI())
	settingsTab.Icon = theme.SettingsIcon()
	tabs := container.NewAppTabs(
		container.NewTabItem("Midweek", config.mGUI(MM)),
		container.NewTabItem("Weekend", config.mGUI(WM)),
		settingsTab,
	)

	w := a.NewWindow("Meeting Downloader")
	w.SetContent(container.NewVBox(tabs))

	w.ShowAndRun()
}
