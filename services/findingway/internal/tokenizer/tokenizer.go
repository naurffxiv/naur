package tokenizer

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/redis/go-redis/v9"
)

type Tokenizer struct {
	PrevParsedPfIds []string
	rdb             *redis.Client
}

type Token struct {
	String string
	Count  int
}

func NowToInt() int {
	startDate := time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC)
	dayCount := time.Since(startDate).Hours() / 24
	return int(dayCount)
}

func splitListingIntoTokens(listing string) ([]string, error) {
	var splitRegex = []string{
		`\|`,   // |
		`\|\|`, // ||
		`\/`,   // /
		`\/\/`, // //
		`->`,   // ->
		`for `, // for
		` `,    // space
		`to `,  // to
		`-`,    // -
		`&`,    // &
		`\+`,   // +
		"\n",   // newline
	}

	fullRegex := strings.Join(splitRegex, `|`)

	re := regexp.MustCompile(fullRegex)

	result := re.Split(listing, -1)

	var resTokens []string

	for _, token := range result {
		if token != "" {
			resTokens = append(resTokens, strings.ToLower(token))
		}
	}

	return resTokens, nil
}

func (t *Tokenizer) Init() {
	t.PrevParsedPfIds = []string{}

	// TODO what is the right way to make sure this is closed?
	t.rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password TODO: pull from settings
		DB:       0,  // use default DB TODO: pull from settings
		Protocol: 2,
	})
}

func (t *Tokenizer) TokenizeListings(listings *ffxiv.Listings) {
	currentDayNumber := NowToInt()

	tokenMap := make(map[string]int)
	pfDescriptions := []string{}

	parsedListingIds := []string{}

	for _, item := range listings.Listings {
		// Currently, parse all PFs for stress testing. In the future, filter to relevant fight
		parsedListingIds = append(parsedListingIds, item.Id)
		if slices.Contains(t.PrevParsedPfIds, item.Id) {
			continue
		}

		pfDescriptions = append(pfDescriptions, item.Description)

		tokenList, _ := splitListingIntoTokens(item.Description)

		for _, token := range tokenList {
			tokenMap[token] += 1
		}
	}

	t.PrevParsedPfIds = parsedListingIds

	ctx := context.Background()

	// store tokens into redis
	todayKey := fmt.Sprintf("tokens:%d", currentDayNumber)
	todayExists, err := t.rdb.Exists(ctx, todayKey).Result()
	if err != nil {
		panic(err)
	}
	for token, count := range tokenMap {
		err := t.rdb.HIncrBy(ctx, todayKey, token, int64(count)).Err()
		if err != nil {
			panic(err)
		}
	}

	if todayExists == 0 {
		t.rdb.Expire(ctx, todayKey, 24*32*time.Hour)
	}

	// Store description list into redis
	descriptionKey := fmt.Sprintf("descriptions:%d", currentDayNumber)
	todayExists, err = t.rdb.Exists(ctx, descriptionKey).Result()
	if err != nil {
		panic(err)
	}

	for _, description := range pfDescriptions {
		err = t.rdb.RPush(ctx, descriptionKey, description).Err()
		if err != nil {
			panic(err)
		}
	}

	if todayExists == 0 {
		t.rdb.Expire(ctx, todayKey, 24*32*time.Hour)
	}
}

func (t *Tokenizer) PrintTokens() {
	// Printing to console for debugging. This can be redirected later
	tokenList := t.GatherTokens(7)
	for _, kv := range tokenList {
		fmt.Printf("%s: %d\n", kv.String, kv.Count)
	}
}

func (t *Tokenizer) GatherTokens(lookback int) []Token {

	tokenSum := make(map[string]int)
	ctx := context.Background()

	todayDayNumber := NowToInt()

	for i := range lookback {
		prevDayNumber := todayDayNumber - i

		getAllResult, err := t.rdb.HGetAll(ctx, fmt.Sprintf("tokens:%d", prevDayNumber)).Result()

		if err != nil {
			panic(err)
		}

		for key, count := range getAllResult {
			intCount, err := strconv.Atoi(count)
			if err != nil {
				panic(err)
			}
			tokenSum[key] += intCount
		}
	}

	var res []Token

	for k, v := range tokenSum {
		res = append(res, Token{k, v})
	}

	// reverse sort, for ease of debugging
	sort.Slice(res, func(i, j int) bool {
		return res[i].Count < res[j].Count
	})

	return res
}

func (t *Tokenizer) CreateCsv(lookback int) {
	// This created CSV locally for debugging. We can point this to something different later
	todayDayNumber := NowToInt()

	fileName := fmt.Sprintf("PfDescriptions_%d.csv", todayDayNumber)
	csvFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)

	csvwriter.Write([]string{"Date", "Description"})
	ctx := context.Background()

	for i := range lookback {
		prevDayNumber := todayDayNumber - i

		getResult, err := t.rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", prevDayNumber), 0, -1).Result()

		if err != nil {
			panic(err)
		}

		for _, description := range getResult {
			fmt.Println(description)
			dateString := time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC).AddDate(0, 0, prevDayNumber).Format(time.DateOnly)
			csvwriter.Write([]string{dateString, strings.ReplaceAll(description, "\n", "")})
		}
	}

	csvwriter.Flush()
	csvFile.Close()
}
