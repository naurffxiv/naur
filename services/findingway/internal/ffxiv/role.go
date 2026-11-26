package ffxiv

type Role int

const (
	DPS Role = iota
	Healer
	Tank
	Empty
)

type Roles struct {
	Roles []Role
}

// roleEmojis maps each combination of open roles to its emoji. Keys are
// order-independent: see key.
var roleEmojis = map[[3]bool]string{
	key(DPS):               "<:dps:1312855997697626185>",
	key(Healer):            "<:healer:1312855996791656449>",
	key(Tank):              "<:tank:1312855995910848573>",
	key(DPS, Healer):       "<:healerdps:1312855994971193384>",
	key(DPS, Tank):         "<:tankdps:1312855993742262424>",
	key(Healer, Tank):      "<:tankhealer:1312855992777576493>",
	key(DPS, Healer, Tank): "<:tankhealerdps:1312855992198758470>",
}

// key records which of DPS, Healer and Tank are present. Empty is ignored.
func key(roles ...Role) [3]bool {
	var k [3]bool
	for _, r := range roles {
		if r >= DPS && r <= Tank {
			k[r] = true
		}
	}
	return k
}

func (rs Roles) Emoji() string {
	if emoji, ok := roleEmojis[key(rs.Roles...)]; ok {
		return emoji
	}
	return "<:AnyRole:1312855998981214268>"
}
