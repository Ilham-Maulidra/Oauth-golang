package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config

var UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

func init() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Error Loading .env file")
	}
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

}

// some random string, random for each request
var oauthStateString = "random"

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleGoogleLogin)
	http.HandleFunc("/callback", handleGoogleCallback)
	fmt.Println("ClientID:", googleOauthConfig.ClientID)
	fmt.Println("ClientSecret:", googleOauthConfig.ClientSecret)
	fmt.Println(http.ListenAndServe(":8080", nil))
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	var htmlIndex = `<html><body>
	<a href="/login">Google Log In</a>
	</body></html>`
	fmt.Fprintln(w, htmlIndex)
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	fmt.Println("=================================")
	fmt.Println("REDIRECT URL:", googleOauthConfig.RedirectURL)
	fmt.Println("OAUTH URL:", url)
	fmt.Println("=================================")

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// state := r.FormValue("state")
	// if state != oauthStateString {
	// 	fmt.Printf("Invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
	// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	// 	return
	// }

	// code := r.FormValue("code")
	// token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	// if err != nil {
	// 	fmt.Printf("Code exchange failed with '%s' \n", err)
	// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	// 	return
	// }
	// resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	// if err != nil {
	// 	panic(err)
	// }
	// defer resp.Body.Close()
	// contents, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Println("Failed to read response body:", err)
	// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	// 	return
	// }
	// fmt.Fprintf(w, "Content: %s\n", contents)

	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(w, "Authorization code not found in request", http.StatusBadRequest)
		return
	}

	token, err := googleOauthConfig.Exchange(
		context.Background(),
		code,
	)

	if err != nil {
		http.Error(w, "Failed to exchange authorization code for token", http.StatusInternalServerError)
		return
	}

	client := googleOauthConfig.Client(
		context.Background(),
		token,
	)

	resp, err := client.Get(
		"https://www.googleapis.com/oauth2/v2/userinfo",
	)

	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusAccepted)
		return
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&UserInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User Info:\nID: %s\nEmail: %s\nVerified Email: %t\nPicture: %s\nName: %s", UserInfo.ID, UserInfo.Email, UserInfo.VerifiedEmail, UserInfo.Picture, UserInfo.Name)

}
