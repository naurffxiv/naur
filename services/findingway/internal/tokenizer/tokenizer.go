package tokenizer

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
)

type Tokenizer struct {
	Tokens          map[int]map[string]int
	PfListingRecord map[int][]string
	PrevParsedPfIds []string
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
	// Key: days since jan 1 1900
	// Value: Map of token string to count of that token's appearances on that day
	t.Tokens = make(map[int]map[string]int)

	// Key: days since jan 1 1900
	// Value: List of PF listings on that day
	t.PfListingRecord = make(map[int][]string)

	t.PrevParsedPfIds = []string{}
}

func (t *Tokenizer) TokenizeListings(listings *ffxiv.Listings) {
	fmt.Println("hello world")

	currentDayNumber := NowToInt()

	todayTokenMap, found := t.Tokens[currentDayNumber]
	if !found {
		todayTokenMap = make(map[string]int)
		t.Tokens[currentDayNumber] = todayTokenMap
	}

	todayPfDescriptionList, found := t.PfListingRecord[currentDayNumber]
	if !found {
		todayPfDescriptionList = []string{}
		t.PfListingRecord[currentDayNumber] = todayPfDescriptionList
	}

	parsedListingIds := []string{}

	for _, item := range listings.Listings {
		// Currently, parse all PFs for stress testing. In the future, filter to relevant fight
		parsedListingIds = append(parsedListingIds, item.Id)
		if slices.Contains(t.PrevParsedPfIds, item.Id) {
			continue
		}

		todayPfDescriptionList = append(todayPfDescriptionList, item.Description)

		tokenList, _ := splitListingIntoTokens(item.Description)

		for _, token := range tokenList {
			todayTokenMap[token] += 1
		}
	}

	t.PfListingRecord[currentDayNumber] = todayPfDescriptionList
	t.PrevParsedPfIds = parsedListingIds
}

func (t *Tokenizer) PrintTokens() {
	// Printing to console for debugging. This can be redirected later
	tokenList := t.GatherTokens(7)
	for _, kv := range tokenList {
		fmt.Printf("%s: %d\n", kv.String, kv.Count)
	}
}

func (t *Tokenizer) GatherTokens(lookback int) []Token {
	fmt.Println("toke it")
	tokenSum := make(map[string]int)

	lookbackDayNumber := NowToInt() - lookback

	for day, dayTokenMap := range t.Tokens {
		if day >= lookbackDayNumber {
			for token, count := range dayTokenMap {
				tokenSum[token] += count
			}
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

	fileName := "PfDescriptions.csv"
	csvFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)

	csvwriter.Write([]string{"Date", "Description"})

	for dateNumber, descriptionList := range t.PfListingRecord {
		dateString := time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC).AddDate(0, 0, dateNumber).Format(time.DateOnly)
		for _, description := range descriptionList {
			// need to remove newline characters that somehow made their way in to descriptions
			csvwriter.Write([]string{dateString, strings.ReplaceAll(description, "\n", "")})
		}
	}

	csvwriter.Flush()
	csvFile.Close()
}

func (t *Tokenizer) ClearOldData() {
	// remove all data older than 30 days

	lastKeptDate := NowToInt() - 30

	// Clear out old Tokens
	oldTokenKeys := []int{}

	for day := range t.Tokens {
		if day < lastKeptDate {
			oldTokenKeys = append(oldTokenKeys, day)
		}
	}

	for key := range oldTokenKeys {
		delete(t.Tokens, key)
	}

	// Clear out old PF listings
	oldPfListKeys := []int{}

	for day := range t.PfListingRecord {
		if day < lastKeptDate {
			oldPfListKeys = append(oldPfListKeys, day)
		}
	}

	for key := range oldPfListKeys {
		delete(t.PfListingRecord, key)
	}
}
