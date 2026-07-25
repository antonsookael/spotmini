package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"spotmini/playback"
)

func launchGUI(accessToken string) {
	a := app.NewWithID("com.antonsookael.spotmini")
	w := a.NewWindow(" ")
	a.Settings().SetTheme(theme.DarkTheme())

	nowPlayingLabel := widget.NewLabel("Loading...")

	// currentSeconds is how far into the song we think we are.
	// It gets corrected every time we actually re-fetch from Spotify.
	var currentSeconds int
	var totalSeconds int
	var currentSong string
	var currentArtist string
	var isCurrentlyPlaying bool

	// formatTime turns a raw second count into "M:SS" style text.
	formatTime := func(totalSecs int) string {
		minutes := totalSecs / 60
		seconds := totalSecs % 60
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	}

	// fetchNowPlaying talks to Spotify and resets our local counter
	// to match reality - fixes drift, and catches song changes/skips.
	fetchNowPlaying := func() {
		state, err := playback.NowPlaying(accessToken)
		if err != nil {
			fyne.Do(func() {
				nowPlayingLabel.SetText("Error loading playback")
			})
			return
		}
		if state.Item.Name == "" {
			currentSong = ""
			fyne.Do(func() {
				nowPlayingLabel.SetText("Nothing playing")
			})
			return
		}

		currentSong = state.Item.Name
		currentArtist = state.Item.Artists[0].Name
		currentSeconds = state.ProgressMs / 1000
		totalSeconds = state.Item.DurationMs / 1000
		isCurrentlyPlaying = state.IsPlaying

		fyne.Do(func() {
			nowPlayingLabel.SetText(fmt.Sprintf("%s by %s — %s/%s",
				currentSong, currentArtist, formatTime(currentSeconds), formatTime(totalSeconds)))
		})
	}

	fetchNowPlaying() // get real data immediately on launch

	// Local ticker: counts up once per second without calling Spotify.
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		secondsSinceSync := 0
		idleSecondsSinceCheck := 0
		for range ticker.C {
			if currentSong == "" {
				// Nothing playing - check every 2 seconds to see if
				// something started, instead of waiting for a button
				// click or the normal 15-second sync.
				idleSecondsSinceCheck++
				if idleSecondsSinceCheck >= 2 {
					idleSecondsSinceCheck = 0
					fetchNowPlaying()
				}
				continue
			}
			idleSecondsSinceCheck = 0

			secondsSinceSync++

			if isCurrentlyPlaying {
				currentSeconds++
				fyne.Do(func() {
					nowPlayingLabel.SetText(fmt.Sprintf("%s by %s — %s/%s",
						currentSong, currentArtist, formatTime(currentSeconds), formatTime(totalSeconds)))
				})
			}

			// Song should have ended - resync immediately rather than
			// waiting for the next 15-second sync, so the display
			// swaps to the new track right away.
			if totalSeconds > 0 && currentSeconds >= totalSeconds {
				secondsSinceSync = 0
				fetchNowPlaying()
				continue
			}

			// Re-sync with the real API every 15 seconds to correct
			// drift and catch song changes that happened via skip/pause.
			if secondsSinceSync >= 15 {
				secondsSinceSync = 0
				fetchNowPlaying()
			}
		}
	}()

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
		fetchNowPlaying()
	})

	previousBtn := widget.NewButton("Previous", func() {
		playback.PreviousTrack(accessToken)
		time.Sleep(300 * time.Millisecond)
		fetchNowPlaying()
	})

	nextBtn := widget.NewButton("Next", func() {
		playback.NextTrack(accessToken)
		time.Sleep(300 * time.Millisecond)
		fetchNowPlaying()
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
		nowPlayingLabel,
		playPauseBtn,
		previousBtn,
		nextBtn,
		shuffleBtn,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}
