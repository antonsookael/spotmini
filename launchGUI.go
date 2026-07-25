package main

import (
	"fmt"
	"spotmini/playback"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func launchGUI(accessToken string) {
	a := app.NewWithID("com.antonsookael.spotmini")
	w := a.NewWindow("Spotify Mini Player")

	playPauseBtn := widget.NewButton("Play/Pause", func() {
		playing, err := playback.IsPlaying(accessToken)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if playing {
			playback.PausePlayback(accessToken)
		} else {
			playback.PlayPlayback(accessToken)
		}
	})

	nextBtn := widget.NewButton("Next", func() {
		playback.NextTrack(accessToken)
	})

	prevBtn := widget.NewButton("Previous", func() {
		playback.PreviousTrack(accessToken)
	})

	shuffleBtn := widget.NewButton("Shuffle", func() {
		shuffled, err := playback.IsShuffled(accessToken)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		playback.ToggleShuffle(accessToken, !shuffled)
	})

	content := container.NewVBox(
		playPauseBtn,
		nextBtn,
		prevBtn,
		shuffleBtn,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}
