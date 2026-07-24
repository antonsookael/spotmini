package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

var clientID, clientSecret string

const redirectURI = "http://127.0.0.1:8888/callback"
const scope = "user-read-playback-state user-modify-playback-state playlist-read-private"
const tokenFile = "token.json"

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: no .env file found - relying on real environment variables instead")
	}

	clientID = os.Getenv("clientID")
	clientSecret = os.Getenv("clientSecret")

	if clientID == "" || clientSecret == "" {
		fmt.Println("Missing CLIENT_ID or CLIENT_SECRET - check your .env file")
		os.Exit(1)
	}
}

func main() {
	loadEnv()
	saved, err := loadToken()
	if err == nil && saved.RefreshToken != "" {
		fmt.Println("Found saved token, refreshing instead of logging in again...")
		newToken, err := refreshToken(saved.RefreshToken)
		if err != nil {
			fmt.Println("Refresh failed, falling back to full login:", err)
			startLogin()
			return
		}
		fmt.Println("Access token:", newToken.AccessToken)
		runCommandLoop(newToken.AccessToken)
		return
	}

	startLogin()
}

func startLogin() {
	authURL := fmt.Sprintf(
		"https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=%s",
		clientID, redirectURI, scope,
	)
	fmt.Println("Open this URL in your browser to log in:")
	fmt.Println(authURL)

	http.HandleFunc("/callback", callbackHandler)
	fmt.Println("\nListening on http://127.0.0.1:8888 ...")
	http.ListenAndServe(":8888", nil)
}

// refreshToken trades a refresh token for a brand new access token,
// without needing the browser at all.
func refreshToken(refresh string) (TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refresh)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	token, err := exchangeForToken(data)
	if err != nil {
		return TokenResponse{}, err
	}

	// Spotify sometimes doesn't return a new refresh_token on refresh -
	// if it didn't, keep using the one we already have.
	if token.RefreshToken == "" {
		token.RefreshToken = refresh
	}

	saveToken(token)
	return token, nil
}

// exchangeForToken is shared by both the initial login and the refresh
// flow - both just POST different form data to the same endpoint.
func exchangeForToken(data url.Values) (TokenResponse, error) {
	resp, err := http.PostForm("https://accounts.spotify.com/api/token", data)
	if err != nil {
		return TokenResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenResponse{}, err
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return TokenResponse{}, fmt.Errorf("parsing response: %w (raw: %s)", err, string(body))
	}

	return token, nil
}

func saveToken(token TokenResponse) {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		fmt.Println("Error saving token:", err)
		return
	}
	if err := os.WriteFile(tokenFile, data, 0644); err != nil {
		fmt.Println("Error writing token file:", err)
	}
}

func loadToken() (TokenResponse, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return TokenResponse{}, err
	}
	var token TokenResponse
	if err := json.Unmarshal(data, &token); err != nil {
		return TokenResponse{}, err
	}
	return token, nil
}
