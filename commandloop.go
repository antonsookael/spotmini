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
			if err := playback.NextTrack(accessToken); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Skipped")
			}

		case "sh":
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
			playing, err := playback.IsPlaying(accessToken)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Is playing:", playing)
			}

		case "h":
			printCommands()

		case "q", "exit":
			fmt.Println("Bye!")
			return

		default:
			fmt.Println("Unknown command. Type 'help' to see options.")
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
}
