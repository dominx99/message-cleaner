package auth_token

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type PersistTeamAttributes struct {
	Team  string
	Token string
}

func (pt PersistTeamAttributes) Persist() (*dynamodb.PutItemOutput, error) {
	db := dynamodb.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

	put := dynamodb.PutItemInput{
		TableName: aws.String("AuthKey"),
		Item: map[string]*dynamodb.AttributeValue{
			"Team": &dynamodb.AttributeValue{
				S: aws.String(pt.Team),
			},
			"AuthToken": &dynamodb.AttributeValue{
				S: aws.String(pt.Token),
			},
		},
	}

	return db.PutItem(&put)
}

type FindAccessTokenAttributes struct {
	Team string
}

func (finder FindAccessTokenAttributes) GetAccessToken() (string, error) {
	db := dynamodb.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(finder.Team),
			},
		},
		KeyConditionExpression: aws.String("Team = :v1"),
		ProjectionExpression:   aws.String("AuthToken"),
		TableName:              aws.String("AuthKey"),
	}

	result, err := db.Query(input)

	return *result.Items[0]["AuthToken"].S, err
}
