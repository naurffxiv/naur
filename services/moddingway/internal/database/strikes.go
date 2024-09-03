package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AddStrikeEntryArgs struct {
	DbUserID   int
	Reason     string
	StrikeTime string
}

func AddStrike(conn *pgxpool.Pool, p AddStrikeEntryArgs) (int, int, error) {
	query := `INSERT INTO strikes (userID, reason, createTimestamp)
	VALUES ($1, $2, $3)
	RETURNING strikeID, (SELECT count(strikeID) from strikes WHERE userID = $1)`

	var strikeID int
	var strikeCount int
	err := conn.QueryRow(
		context.Background(),
		query,
		p.DbUserID,
		p.Reason,
		p.StrikeTime,
	).Scan(&strikeID, &strikeCount)

	if err != nil {
		return -1, -1, err
	}

	// currently added strike is not counted when current strike count is returned
	// Adding one here makes sure correct value is returned
	return strikeID, strikeCount + 1, nil
}

type Strike struct {
	StrikeID       int
	Reason         string
	DiscordUserID  string
	DiscordGuildID string
}

func ListUserStrikes(conn *pgxpool.Pool, userID int) ([]Strike, error) {
	query := `SELECT s.strikeID, s.reason, u.discordUserID, u.discordGuildID
	FROM strikes s
	JOIN users u ON s.userID = u.userID
	WHERE s.userID = $1;`

	rows, err := conn.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strikes []Strike
	for rows.Next() {
		currRow := Strike{}
		err := rows.Scan(
			&currRow.StrikeID,
			&currRow.Reason,
			&currRow.DiscordUserID,
			&currRow.DiscordGuildID,
		)
		if err != nil {
			return nil, err
		}
		strikes = append(strikes, currRow)
	}

	return strikes, nil
}

func RemoveStrike(conn *pgxpool.Pool, strikeID int) error {
	query := `DELETE FROM strikes WHERE strikeid = $1`

	_, err := conn.Query(
		context.Background(),
		query,
		strikeID,
	)

	return err
}

func ClearStrikesForUser(conn *pgxpool.Pool, userID int) error {
	query := `DELETE FROM strikes WHERE userID = $1`

	_, err := conn.Query(
		context.Background(),
		query,
		userID,
	)

	return err
}
