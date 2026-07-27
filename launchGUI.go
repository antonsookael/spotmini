package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"golang.design/x/hotkey"

	"spotmini/playback"
)

func launchGUI(accessToken string) {
	a := app.NewWithID("com.antonsookael.spotmini")
	w := a.NewWindow("spotmini")
	a.Settings().SetTheme(theme.DarkTheme())

	nowPlayingLabel := widget.NewLabel("Loading...")

	var currentSeconds int
	var totalSeconds int
	var currentSong string
	var currentArtist string
	var isCurrentlyPlaying bool

	formatTime := func(totalSecs int) string {
		minutes := totalSecs / 60
		seconds := totalSecs % 60
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	}

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
			nowPlayingLabel.SetText(fmt.Sprintf("%s - %s  %s/%s",
				currentSong, currentArtist, formatTime(currentSeconds), formatTime(totalSeconds)))
		})
	}

	fetchNowPlaying()

	// Local ticker, unchanged from before.
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		secondsSinceSync := 0
		idleSecondsSinceCheck := 0
		for range ticker.C {
			if currentSong == "" {
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
					nowPlayingLabel.SetText(fmt.Sprintf("%s - %s  %s/%s",
						currentSong, currentArtist, formatTime(currentSeconds), formatTime(totalSeconds)))
				})
			}

			if totalSeconds > 0 && currentSeconds >= totalSeconds {
				secondsSinceSync = 0
				fetchNowPlaying()
				continue
			}

			if secondsSinceSync >= 15 {
				secondsSinceSync = 0
				fetchNowPlaying()
			}
		}
	}()

	// These used to be inline button click handlers - now they're
	// named functions so both a hotkey AND (optionally) a button
	// can trigger the same action without duplicating logic.
	doPlayPause := func() {
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
	}

	doPrevious := func() {
		playback.PreviousTrack(accessToken)
		time.Sleep(300 * time.Millisecond)
		fetchNowPlaying()
	}

	doNext := func() {
		playback.NextTrack(accessToken)
		time.Sleep(300 * time.Millisecond)
		fetchNowPlaying()
	}

	doShuffle := func() {
		shuffled, err := playback.IsShuffled(accessToken)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		playback.ToggleShuffle(accessToken, !shuffled)
	}

	registerHotkeys(doPlayPause, doPrevious, doNext, doShuffle)

	// Minimal strip: just the now-playing label, no buttons.
	// All control happens via global hotkeys now.
	w.SetContent(nowPlayingLabel)
	w.Resize(fyne.NewSize(320, 40))
	w.SetFixedSize(true)

	w.ShowAndRun()
}

// registerHotkeys sets up global keyboard shortcuts that work even
// when the window isn't focused - this is the whole point of a
// minimal overlay: you never need to click into it.
//
// Defaults (adjust to taste):
//
//	Ctrl+Alt+Space -> Play/Pause
//	Ctrl+Alt+Left  -> Previous
//	Ctrl+Alt+Right -> Next
//	Ctrl+Alt+S     -> Shuffle
func registerHotkeys(onPlayPause, onPrevious, onNext, onShuffle func()) {
	playPauseHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}, hotkey.KeySpace)
	previousHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}, hotkey.KeyLeft)
	nextHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}, hotkey.KeyRight)
	shuffleHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}, hotkey.KeyS)

	registerOne(playPauseHK, "Play/Pause", onPlayPause)
	registerOne(previousHK, "Previous", onPrevious)
	registerOne(nextHK, "Next", onNext)
	registerOne(shuffleHK, "Shuffle", onShuffle)
}

// registerOne registers a single hotkey and spins up a goroutine that
// listens for it firing, forever, calling the given function each time.
func registerOne(hk *hotkey.Hotkey, name string, action func()) {
	if err := hk.Register(); err != nil {
		fmt.Println("Failed to register hotkey for", name, ":", err)
		return
	}

	go func() {
		for range hk.Keydown() {
			action()
		}
	}()
}
