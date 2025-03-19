package auth

import (
	"encoding/json"
	"fmt"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

var (
	googleOauthConfig *oauth2.Config
)

func InitGoogleOAuth() {
	googleOauthConfig = &oauth2.Config{
		ClientID:     config.GetGoogleClientID(),
		ClientSecret: config.GetGoogleClientSecret(),
		RedirectURL:  config.GetGoogleRedirectURL(),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func GetGoogleOAuthConfig() *oauth2.Config {
	return googleOauthConfig
}

func GetGoogleUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
	client := googleOauthConfig.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed decoding user info: %s", err.Error())
	}

	return &userInfo, nil
}
