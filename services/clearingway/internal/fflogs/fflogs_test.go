package fflogs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func NewTestFFLogs(client GraphQLClientInterface) *FFLogs {
	return &FFLogs{
		OAuthToken: &oauth2.Token{
			Expiry: time.Now().Add(time.Hour),
		},
		GraphQLClient: client,
	}
}

func TestCheckCharacterClear_TotalKillsGreaterThanZero(t *testing.T) {
	ffl := NewTestFFLogs(&MockGraphQLClient{
		QueryFunc: func(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error {
			query := q.(*CharacterRankingsQuery)
			rankings := map[string]interface{}{
				"totalKills": 5.0,
				"bestSpeed":  120.0,
			}
			rankingsBytes, _ := json.Marshal(rankings)
			query.CharacterData.Character.EncounterRankings = rankingsBytes
			query.CharacterData.Character.LodestoneID = 12345
			return nil
		},
	})

	ctx := context.Background()
	result, err := ffl.CheckCharacterClear(ctx, 12345, 1000)

	assert.NoError(t, err)
	assert.True(t, result)
}

func TestCheckCharacterClear_TotalKillsEqualToZero(t *testing.T) {
	ffl := NewTestFFLogs(&MockGraphQLClient{
		QueryFunc: func(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error {
			query := q.(*CharacterRankingsQuery)
			rankings := map[string]interface{}{
				"totalKills": 0.0,
				"bestSpeed":  120.0,
			}
			rankingsBytes, _ := json.Marshal(rankings)
			query.CharacterData.Character.EncounterRankings = rankingsBytes
			return nil
		},
	})

	ctx := context.Background()
	result, err := ffl.CheckCharacterClear(ctx, 12345, 1000)

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestCheckCharacterClear_TotalKillsMissing(t *testing.T) {
	ffl := NewTestFFLogs(&MockGraphQLClient{
		QueryFunc: func(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error {
			query := q.(*CharacterRankingsQuery)
			rankings := map[string]interface{}{
				"bestSpeed": 120.0,
			}
			rankingsBytes, _ := json.Marshal(rankings)
			query.CharacterData.Character.EncounterRankings = rankingsBytes
			return nil
		},
	})

	ctx := context.Background()
	result, err := ffl.CheckCharacterClear(ctx, 12345, 1000)

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestCheckCharacterClear_InvalidJSON(t *testing.T) {
	ffl := NewTestFFLogs(&MockGraphQLClient{
		QueryFunc: func(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error {
			query := q.(*CharacterRankingsQuery)
			query.CharacterData.Character.EncounterRankings = json.RawMessage("invalid json")
			return nil
		},
	})

	ctx := context.Background()
	result, err := ffl.CheckCharacterClear(ctx, 12345, 1000)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
	assert.False(t, result)
}

func TestCheckCharacterClear_GraphQLError(t *testing.T) {
	expectedErr := errors.New("graphQL query failed")
	ffl := NewTestFFLogs(&MockGraphQLClient{
		QueryFunc: func(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error {
			return expectedErr
		},
	})

	ctx := context.Background()
	result, err := ffl.CheckCharacterClear(ctx, 12345, 1000)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.False(t, result)
}

func TestCheckCharacterClear_VariablesPassedCorrectly(t *testing.T) {
	var capturedLodestoneId int64
	var capturedEncounterId int64

	ffl := NewTestFFLogs(&MockGraphQLClient{
		QueryFunc: func(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...graphql.Option) error {
			query := q.(*CharacterRankingsQuery)
			rankings := map[string]interface{}{
				"totalKills": 1.0,
			}
			rankingsBytes, _ := json.Marshal(rankings)
			query.CharacterData.Character.EncounterRankings = rankingsBytes

			capturedLodestoneId = variables["lodestoneId"].(int64)
			capturedEncounterId = variables["encounterId"].(int64)
			return nil
		},
	})

	ctx := context.Background()
	_, err := ffl.CheckCharacterClear(ctx, 99999, 88888)

	assert.NoError(t, err)
	assert.Equal(t, int64(99999), capturedLodestoneId)
	assert.Equal(t, int64(88888), capturedEncounterId)
}
