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

	json.NewEncoder(w).Encode(req.Form)

	api, err := getApi(req.Form.Get("team_id"))

	if err != nil {
		json.NewEncoder(w).Encode("Cannot find access token to your application.")
		return
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

	if !result.Ok {
		json.NewEncoder(w).Encode("Something went wrong.")
		return
	}

	d := auth_token.DeleteAccessTokenAttributes{
		Team: result.TeamID,
	}

	var deleteError error = d.DeleteAccessToken()

	if deleteError != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	p := auth_token.PersistTeamAttributes{
		Team:  result.TeamID,
		Token: result.AccessToken,
	}

	_, persistError := p.Persist()

	if persistError != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	json.NewEncoder(w).Encode("success")
}

func setStatus(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	api, err := getApi(req.Form.Get("team_id"))

	if err != nil {
		json.NewEncoder(w).Encode("Cannot find access token to your application.")
	}

	s := status_repo.Status{
		Name: req.Form.Get("text"),
		User: req.Form.Get("user_id"),
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
