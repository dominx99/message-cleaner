package message_repo

import (
	"github.com/nlopes/slack"
)

type Messages struct {
	ChannelID   string
	ChannelName string
}

func (m *Messages) Load(api *slack.Client, messages *[]slack.Message) error {
	var messageError error

	switch m.ChannelName {
	case "privategroup":
		history, err := api.GetGroupHistory(
			m.ChannelID,
			slack.NewHistoryParameters(),
		)

		*messages = history.Messages
		messageError = err
	case "directmessage":
		params := slack.GetConversationHistoryParameters{
			ChannelID: m.ChannelID,
			Latest:    slack.DEFAULT_HISTORY_LATEST,
			Oldest:    slack.DEFAULT_HISTORY_OLDEST,
			Inclusive: slack.DEFAULT_HISTORY_INCLUSIVE,
			Limit:     100,
		}

		history, err := api.GetConversationHistory(&params)

		*messages = history.Messages
		messageError = err
	default:
		history, err := api.GetChannelHistory(
			m.ChannelID,
			slack.NewHistoryParameters(),
		)

		*messages = history.Messages
		messageError = err
	}

	return messageError
}

func (m *Messages) BulkDelete(api *slack.Client, messages []slack.Message) error {
	var resultErr error

	for i := 0; i < len(messages); i++ {
		if messages[i].IsStarred {
			continue
		}

		_, _, err := api.DeleteMessage(
			m.ChannelID,
			messages[i].Timestamp,
		)

		if err != nil {
			resultErr = err
		}
	}

	return resultErr
}
