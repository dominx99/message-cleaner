package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/nlopes/slack"
)

type AuthKey struct {
	Token  string
	Team string
}

func persistTeam(w http.ResponseWriter, token, team string) {
	db := dynamodb.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

    del := dynamodb.DeleteItemInput{
        TableName: aws.String("AuthKey"),
        Key: map[string]*dynamodb.AttributeValue {
            "Team": &dynamodb.AttributeValue {
                S: aws.String(team),
            },
            "AuthToken": &dynamodb.AttributeValue {
                S: aws.String(token),
            },
        },
    }

    resd, err := db.DeleteItem(&del)

    if err != nil {
        json.NewEncoder(w).Encode(err.Error())
    } else {
        json.NewEncoder(w).Encode(resd)
    }

    put := dynamodb.PutItemInput{
        TableName: aws.String("AuthKey"),
        Item: map[string]*dynamodb.AttributeValue {
            "Team": &dynamodb.AttributeValue {
                S: aws.String(team),
            },
            "AuthToken": &dynamodb.AttributeValue {
                S: aws.String(token),
            },
        },
    }

    in, err := db.PutItem(&put)

    if err != nil {
        json.NewEncoder(w).Encode(err.Error())
        return
    }

    json.NewEncoder(w).Encode(in)
}

func findTokenByTeam(team string) *string {
	db := dynamodb.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

    input := &dynamodb.QueryInput{
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":v1": {
                S: aws.String(team),
            },
        },
        KeyConditionExpression: aws.String("Team = :v1"),
        ProjectionExpression:   aws.String("AuthToken"),
        TableName:              aws.String("AuthKey"),
    }

    result, err := db.Query(input)

    if err != nil {
        fmt.Println(err.Error())
    }

    return result.Items[0]["AuthToken"].S
}

func clear(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

    var token *string = findTokenByTeam(req.Form.Get("team_id"))

	api := slack.New(*token)

	chis, err := api.GetChannelHistory(
		req.Form.Get("channel_id"),
		slack.NewHistoryParameters(),
	)

    var messages []slack.Message

    if err == nil {
        messages = chis.Messages
    } else {
        ghis, err := api.GetGroupHistory(
            req.Form.Get("channel_id"),
            slack.NewHistoryParameters(),
        )

        if err == nil {
            messages = ghis.Messages
        } else {
            params := slack.GetConversationHistoryParameters{
                ChannelID: req.Form.Get("channel_id"),
                Latest:    slack.DEFAULT_HISTORY_LATEST,
                Oldest:    slack.DEFAULT_HISTORY_OLDEST,
                Inclusive: slack.DEFAULT_HISTORY_INCLUSIVE,
                Limit: 100,
            }

            cohis, err := api.GetConversationHistory(&params)

            if err == nil {
                messages = cohis.Messages
            } else {
                json.NewEncoder(w).Encode(err.Error())
            }
        }
    }

	for i := 0; i < len(messages); i++ {
        _, _, err := api.DeleteMessage(
			req.Form.Get("channel_id"),
			messages[i].Timestamp,
		)

        if err != nil {
		    json.NewEncoder(w).Encode(err.Error())
        }
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
	}

	var result slack.OAuthResponse

	json.NewDecoder(res.Body).Decode(&result)

	persistTeam(w, result.AccessToken, result.TeamID)

	if res.StatusCode == 200 {
		json.NewEncoder(w).Encode("success")
	} else {
		json.NewEncoder(w).Encode("failed")
	}
}

func main() {
	http.HandleFunc("/", clear)
	http.HandleFunc("/redirect", redirect)

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
