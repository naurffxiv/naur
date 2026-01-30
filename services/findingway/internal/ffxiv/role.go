package ffxiv

import (
	"reflect"
)

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

func (rs Roles) Emoji() string {
	if reflect.DeepEqual(rs.Roles, []Role{DPS}) {
		return "<:dps:1312855997697626185>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{Healer}) {
		return "<:healer:1312855996791656449>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{Tank}) {
		return "<:tank:1312855995910848573>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{DPS, Healer}) {
		return "<:healerdps:1312855994971193384>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{DPS, Tank}) {
		return "<:tankdps:1312855993742262424>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{Healer, Tank}) {
		return "<:tankhealer:1312855992777576493>"
	}

	if reflect.DeepEqual(rs.Roles, []Role{Healer, Tank, DPS}) {
		return "<:tankhealerdps:1312855992198758470>"
	}

	return "<:AnyRole:1312855998981214268>"
}
