package main

import (
	"encoding/json"
	"net/http"
	"os"

	slack "example.com/api"
	auth_token "example.com/auth"
)

func clear(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	accessTokenFinder := auth_token.FindAccessTokenAttributes{
		Team: req.Form.Get("team_id"),
	}

	token, err := accessTokenFinder.GetAccessToken()

	if err != nil {
		json.NewEncoder(w).Encode("Cannot find access token to your application.")
		return
	}

	api := slack.Api{
		Token: token,
	}

	loadParams := slack.HistoryParameters{
		ChannelName: req.Form.Get("channel_name"),
		ChannelID:   req.Form.Get("channel_id"),
	}

	var history slack.History

	loadError := api.GetChannelHistory(loadParams, &history)

	if loadError != nil {
		json.NewEncoder(w).Encode(loadError.Error())
		return
	}

	deleteParams := slack.DeleteMessageHistoryParameters{
		ChannelName: req.Form.Get("channel_name"),
		ChannelID:   req.Form.Get("channel_id"),
	}

	deleteError := api.DeleteNotStarredMessages(deleteParams, history)

	if deleteError != nil {
		json.NewEncoder(w).Encode(deleteError.Error())
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

func main() {
	http.HandleFunc("/", clear)
	http.HandleFunc("/redirect", redirect)

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
