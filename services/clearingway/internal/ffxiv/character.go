package ffxiv

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ErrInvalidWorld = errors.New("invalid world")
	ErrInvalidName  = errors.New("invalid character name")
)

const MinNameLength = 2

type Character struct {
	FirstName   string
	LastName    string
	World       string
	LodestoneID int64
}

// NewCharacter - Validates and normalises the provided name and world, returning a new Character
func NewCharacter(firstName, lastName, world string) (*Character, error) {
	canonicalWorld, ok := FindWorld(world)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrInvalidWorld, world)
	}

	firstName = normaliseName(firstName)
	lastName = normaliseName(lastName)
	if len([]rune(firstName)) < MinNameLength {
		return nil, fmt.Errorf("%w: first name must be at least %d characters", ErrInvalidName, MinNameLength)
	}
	if len([]rune(lastName)) < MinNameLength {
		return nil, fmt.Errorf("%w: last name must be at least %d characters", ErrInvalidName, MinNameLength)
	}

	return &Character{
		FirstName: firstName,
		LastName:  lastName,
		World:     canonicalWorld,
	}, nil
}

// Name - Returns the character's full name, i.e. "Firstname Lastname"
func (c *Character) Name() string {
	return c.FirstName + " " + c.LastName
}

// normaliseName - Trims, title-cases, and replaces typographic apostrophes in a name
func normaliseName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "’", "'")
	return cases.Title(language.AmericanEnglish).String(name)
}
