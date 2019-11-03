package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/nlopes/slack"

	auth_token "example.com/auth"
	message_repo "example.com/messages"
	status_repo "example.com/status"
)

func getApi(teamId string) (*slack.Client, error) {
	accessTokenFinder := auth_token.FindAccessTokenAttributes{
		Team: teamId,
	}

	token, err := accessTokenFinder.GetAccessToken()

	return slack.New(token), err
}

func clear(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	api, err := getApi(req.Form.Get("team_id"))

	if err != nil {
		json.NewEncoder(w).Encode("Cannot find access token to your application.")
	}

	m := message_repo.Messages{
		ChannelID:   req.Form.Get("channel_id"),
		ChannelName: req.Form.Get("channel_name"),
	}

	var messages []slack.Message

	loadError := m.Load(api, &messages)
	deleteError := m.BulkDelete(api, messages)

	if loadError != nil {
		json.NewEncoder(w).Encode(loadError)
	}

	if deleteError != nil {
		json.NewEncoder(w).Encode(deleteError)
	}
}

func redirect(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()

	res, err := http.Get(
		"https://slack.com/api/oauth.access?code=" +
			query.Get("code") +
			"&client_id=" + os.Getenv("SLACK_CLIENT_ID") +
			"&client_secret=" + os.Getenv("SLACK_CLIENT_SECRET") +
			"&redirect_uri=" + os.Getenv("SLACK_REDIRECT"),
	)

	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	var result slack.OAuthResponse

	json.NewDecoder(res.Body).Decode(&result)

	p := auth_token.PersistTeamAttributes{
		Team:  result.AccessToken,
		Token: result.TeamID,
	}

	_, err = p.Persist()

	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if res.StatusCode == 200 {
		json.NewEncoder(w).Encode("success")
	} else {
		json.NewEncoder(w).Encode("failed")
	}
}

func setStatus(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	api, err := getApi(req.Form.Get("team_id"))

	if err != nil {
		json.NewEncoder(w).Encode("Cannot find access token to your application.")
	}

    s := status_repo.Status{
        Name: req.Form.Get("text"),
    }
    err = s.Set(api)

	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
	}
}

func main() {
	http.HandleFunc("/", clear)
	http.HandleFunc("/status", setStatus)
	http.HandleFunc("/redirect", redirect)

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
