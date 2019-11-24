package message_repo

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/nlopes/slack"
)

type Messages struct {
	Token       string
	ChannelID   string
	ChannelName string
}

type History struct {
    Messages []struct {
        Ts string `json:"ts"`
        Text string `json:"text"`
        IsStarred bool `json:"is_stared"`
    } `json:"messages"`
}

func (m *Messages) Load(api *slack.Client, history *History) error {
	var messageError error

	switch m.ChannelName {
	case "privategroup":
		var queryString string = "?token=" + m.Token + "&channel=" + m.ChannelID

		res, err := http.Get("https://slack.com/api/groups.history" + queryString)

		messageError = err

        json.NewDecoder(res.Body).Decode(history)

		// history, err := api.GetGroupHistory(
		// 	m.ChannelID,
		// 	slack.NewHistoryParameters(),
		// )

		// *messages = history.Messages

		// messageError = err
	case "directmessage":
		var queryString string = "?token=" + m.Token + "&channel=" + m.ChannelID

		res, err := http.Get("https://slack.com/api/conversations.history" + queryString)

		messageError = err

        json.NewDecoder(res.Body).Decode(history)

		// params := slack.GetConversationHistoryParameters{
		// 	ChannelID: m.ChannelID,
		// 	Latest:    slack.DEFAULT_HISTORY_LATEST,
		// 	Oldest:    slack.DEFAULT_HISTORY_OLDEST,
		// 	Inclusive: slack.DEFAULT_HISTORY_INCLUSIVE,
		// 	Limit:     100,
		// }

		// history, err := api.GetConversationHistory(&params)

		// *messages = history.Messages
		// messageError = err
	default:
		var queryString string = "?token=" + m.Token + "&channel=" + m.ChannelID

		res, err := http.Get("https://slack.com/api/channels.history" + queryString)

		messageError = err

        json.NewDecoder(res.Body).Decode(history)

		// history, err := api.GetChannelHistory(
		// 	m.ChannelID,
		// 	slack.NewHistoryParameters(),
		// )

		// *messages = history.Messages
		// messageError = err
	}

	return messageError
}

func (m *Messages) BulkDelete(api *slack.Client, history History) error {
	var resultErr error

	for i := 0; i < len(history.Messages); i++ {
		if history.Messages[i].IsStarred {
			continue
		}

		_, _, _, err := api.SendMessageContext(
			context.Background(),
			m.ChannelID,
			slack.MsgOptionDelete(history.Messages[i].Ts),
			slack.MsgOptionAsUser(false),
		)

		if err != nil {
			resultErr = err
		}
	}

	return resultErr
}
