package fflogs

import (
	"clearingway/internal/env"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type GraphQLClientInterface interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error
}

type FFLogs struct {
	ClientId      string
	ClientSecret  string
	OAuthToken    *oauth2.Token
	GraphQLClient GraphQLClientInterface
}

// Init - Initializes the FFLogs struct by generating an OAuth token and setting up the GraphQL client
func Init(ctx context.Context, env *env.Env) *FFLogs {
	ffl := &FFLogs{
		ClientId:     env.FFLOGS_CLIENT_ID,
		ClientSecret: env.FFLOGS_CLIENT_SECRET,
	}

	// Generate initial OAuth token and set up GraphQL client
	ffl.RefreshOAuthClient(ctx)
	return ffl
}

// generateOAuthToken - Generates an OAuth token using the client credentials flow and stores it in the FFLogs struct
func (ffl *FFLogs) generateOAuthToken(ctx context.Context) *oauth2.Token {
	config := &clientcredentials.Config{
		ClientID:     ffl.ClientId,
		ClientSecret: ffl.ClientSecret,
		TokenURL:     "https://www.fflogs.com/oauth/token",
	}

	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("Failed to generate OAuth token: %v", err)
	}

	ffl.OAuthToken = token
	return token
}

// newGraphQLClient - Initializes a new GraphQL client with the current OAuth token
func (ffl *FFLogs) newGraphQLClient(ctx context.Context) *graphql.Client {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(ffl.OAuthToken))
	return graphql.NewClient("https://www.fflogs.com/api/v2/client", httpClient)
}

// RefreshOAuthClient - Refreshes the OAuth token and updates the GraphQL client with the new token
func (ffl *FFLogs) RefreshOAuthClient(ctx context.Context) {
	ffl.generateOAuthToken(ctx)
	ffl.GraphQLClient = ffl.newGraphQLClient(ctx)
}

// CheckCharacterClear - Fetches character rankings for a given lodestoneId and encounterId.
// Returns false if "totalKills" is not present in the EncounterRankings or 0, true otherwise
func (ffl *FFLogs) CheckCharacterClear(ctx context.Context, lodestoneId, encounterId int64) (bool, error) {
	// Check if token is expired and refresh if necessary
	if ffl.OAuthToken == nil || ffl.OAuthToken.Expiry.Before(time.Now()) {
		ffl.RefreshOAuthClient(ctx)
	}

	// Define the GraphQL query and variables and execute
	var q CharacterRankingsQuery
	if err := ffl.GraphQLClient.Query(ctx, &q, map[string]interface{}{
		"lodestoneId": lodestoneId,
		"encounterId": encounterId,
	}); err != nil {
		return false, err
	}

	// Unmarshal the EncounterRankings JSON string into a map to check for totalKills
	var erData map[string]interface{}
	if err := json.Unmarshal(q.CharacterData.Character.EncounterRankings, &erData); err != nil {
		return false, err
	}

	// Check if totalKills is present and greater than 0
	totalKills, ok := erData["totalKills"].(float64) // JSON numbers are unmarshaled as float64
	if !ok || totalKills == 0 {
		return false, nil
	}

	return true, nil
}
