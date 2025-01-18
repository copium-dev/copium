package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"jtracker-backend/config"
	"net/http"
)

var oauthStateString = "random-state-token" // Example state token; ensure it's consistent across requests

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := config.GoogleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Check the state parameter
	state := r.FormValue("state")
	if state != oauthStateString {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := config.GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Code exchange failed", http.StatusInternalServerError)
		return
	}

	// Exchange the authorization code for an access token
	client := config.GoogleOauthConfig.Client(context.Background(), token)
	userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer userInfoResp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	userInfoJSON, _ := json.MarshalIndent(userInfo, "", "  ")
	fmt.Fprintf(w, "User Info: %s", userInfoJSON)
}
