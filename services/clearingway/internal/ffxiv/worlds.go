package ffxiv

import "strings"

var worlds = []string{
	// North America
	"Adamantoise", "Cactuar", "Faerie", "Gilgamesh", "Jenova", "Midgardsormr", "Sargatanas", "Siren",
	"Balmung", "Brynhildr", "Coeurl", "Diabolos", "Goblin", "Malboro", "Mateus", "Zalera",
	"Cuchulainn", "Golem", "Halicarnassus", "Kraken", "Maduin", "Marilith", "Rafflesia", "Seraph",
	"Behemoth", "Excalibur", "Exodus", "Famfrit", "Hyperion", "Lamia", "Leviathan", "Ultros",
	// Europe
	"Cerberus", "Louisoix", "Moogle", "Omega", "Phantom", "Ragnarok", "Sagittarius", "Spriggan",
	"Alpha", "Lich", "Odin", "Phoenix", "Raiden", "Shiva", "Twintania", "Zodiark",
	"Innocence", "Pixie", "Titania", "Tycoon",
	// Oceania
	"Bismarck", "Ravana", "Sephirot", "Sophia", "Zurvan",
	// Japan
	"Aegis", "Atomos", "Carbuncle", "Garuda", "Gungnir", "Kujata", "Tonberry", "Typhon",
	"Alexander", "Bahamut", "Durandal", "Fenrir", "Ifrit", "Ridill", "Tiamat", "Ultima",
	"Anima", "Asura", "Chocobo", "Hades", "Ixion", "Masamune", "Pandaemonium", "Titan",
	"Belias", "Mandragora", "Ramuh", "Shinryu", "Unicorn", "Valefor", "Yojimbo", "Zeromus",
}

// AllWorlds - Returns the names of all worlds
func AllWorlds() []string {
	return append([]string(nil), worlds...)
}

// FindWorld - Case-insensitively looks up a world and returns its proper name
func FindWorld(name string) (string, bool) {
	name = strings.TrimSpace(name)
	for _, world := range worlds {
		if strings.EqualFold(world, name) {
			return world, true
		}
	}
	return "", false
}

// IsWorld - Returns true if the provided name is a valid world
func IsWorld(name string) bool {
	_, ok := FindWorld(name)
	return ok
}
