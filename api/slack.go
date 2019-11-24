package slack

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type SlackResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type OAuthResponseIncomingWebhook struct {
	URL              string `json:"url"`
	Channel          string `json:"channel"`
	ChannelID        string `json:"channel_id,omitempty"`
	ConfigurationURL string `json:"configuration_url"`
}

type OAuthResponseBot struct {
	BotUserID      string `json:"bot_user_id"`
	BotAccessToken string `json:"bot_access_token"`
}

type OAuthResponse struct {
	AccessToken     string                       `json:"access_token"`
	Scope           string                       `json:"scope"`
	TeamName        string                       `json:"team_name"`
	TeamID          string                       `json:"team_id"`
	IncomingWebhook OAuthResponseIncomingWebhook `json:"incoming_webhook"`
	Bot             OAuthResponseBot             `json:"bot"`
	UserID          string                       `json:"user_id,omitempty"`
	SlackResponse
}

type Api struct {
	Token string
}

type HistoryParameters struct {
	ChannelName string
	ChannelID   string
}

type History struct {
    Messages []struct {
        Timestamp string `json:"ts"`
        Text string `json:"text"`
        IsStarred bool `json:"is_stared"`
    } `json:"messages"`
}

func (api *Api) GetChannelHistory(params HistoryParameters, history *History) error {
    var action string

	switch params.ChannelName {
	case "privategroup":
        action = "groups.history"
	case "directmessage":
        action = "conversations.history"
	default:
        action = "channels.history"
	}

    var queryString string = "?token=" + api.Token + "&channel=" + params.ChannelID

    res, err := http.Get("https://slack.com/api/" + action + queryString)

    json.NewDecoder(res.Body).Decode(history)

	return err
}

type DeleteMessageHistoryParameters struct {
    ChannelID string
    ChannelName string
}

func (api *Api) DeleteNotStarredMessages(params DeleteMessageHistoryParameters, history History) error {
	var resultErr error

	for i := 0; i < len(history.Messages); i++ {
		if history.Messages[i].IsStarred {
			continue
		}

		_, err := http.PostForm("https://slack.com/api/chat.delete", url.Values{
            "token": {api.Token},
            "channel": {params.ChannelID},
            "ts": {history.Messages[i].Timestamp},
        })

		if err != nil {
			resultErr = err
		}
	}

	return resultErr
}
