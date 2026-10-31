package fflogs

import "encoding/json"

type CharacterRankingsQuery struct {
	CharacterData struct {
		Character struct {
			EncounterRankings json.RawMessage `graphql:"encounterRankings(encounterID: $encounterId)"`
			LodestoneID       int             `graphql:"lodestoneID"`
		} `graphql:"character(lodestoneID: $lodestoneId)"`
	} `graphql:"characterData"`
}
