package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"spotmini/playback"
)

func runCommandLoop(accessToken string) {
	scanner := bufio.NewScanner(os.Stdin)
	printCommands()

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())

		switch command {
		case "":
			fmt.Print("\033[H\033[2J\033[3J")
			printCommands()
			if playing, err := playback.IsPlaying(accessToken); err != nil {
				fmt.Println("Error:", err)
			} else if playing {
				if err := playback.PausePlayback(accessToken); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Paused")
				}
			} else {
				if err := playback.PlayPlayback(accessToken); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Playing")
				}
			}

		case "n":
			fmt.Print("\033[H\033[2J\033[3J")
			printCommands()
			if err := playback.NextTrack(accessToken); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Skipped")
			}

		case "sh":
			fmt.Print("\033[H\033[2J\033[3J")
			printCommands()
			if shuffled, err := playback.IsShuffled(accessToken); err != nil {
				fmt.Println("Error:", err)
			} else if shuffled {
				if err := playback.ToggleShuffle(accessToken, false); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Shuffle off")
				}
			} else {
				if err := playback.ToggleShuffle(accessToken, true); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Shuffle on")
				}
			}

		case "s":
			fmt.Print("\033[H\033[2J\033[3J")
			printCommands()
			state, err := playback.NowPlaying(accessToken)
			if err != nil {
				fmt.Println("Error:", err)
			} else if state.Item.Name == "" {
				fmt.Println("Nothing playing")
			} else {
				minutes := state.ProgressMs / 1000 / 60
				seconds := (state.ProgressMs / 1000) % 60
				fmt.Printf("%s by %s — %d:%02d\n", state.Item.Name, state.Item.Artists[0].Name, minutes, seconds)
			}

		case "h":
			fmt.Print("\033[H\033[2J\033[3J")
			printCommands()

		case "q", "exit":
			fmt.Print("\033[H\033[2J\033[3J")
			fmt.Println("Bye!")
			return

		default:
			fmt.Print("\033[H\033[2J\033[3J")
			printCommands()
			fmt.Println("Unknown command. Type 'h' to see options.")
		}
	}
}

func printCommands() {
	fmt.Println("Commands:")
	fmt.Println("  n - Next track")
	fmt.Println("  sh - Toggle shuffle")
	fmt.Println("  s - Show status")
	fmt.Println("  h - Show help")
	fmt.Println("  q - Quit")
	fmt.Println("-------------------------")
}
