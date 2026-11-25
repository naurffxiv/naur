package lodestone

import (
	"fmt"
	"hash/adler32"
)

// OwnershipCode - Returns the code a user must place in their Lodestone bio to prove ownership of a character
func OwnershipCode(discordId string) string {
	return fmt.Sprintf("clearingway-%d", adler32.Checksum([]byte(discordId)))
}
