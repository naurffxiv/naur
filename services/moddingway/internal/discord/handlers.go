package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	// I am too dumb and tired to figure out how to set permission restrictions for the roles without these vars.. sorry..
	kickPermission    int64 = discordgo.PermissionKickMembers
	mutePermission    int64 = discordgo.PermissionModerateMembers
	banPermission     int64 = discordgo.PermissionBanMembers
	channelPermission int64 = discordgo.PermissionManageChannels
	messagePermission int64 = discordgo.PermissionManageMessages
)

func (d *Discord) AddCommands(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("Initializing Discord...\n")

	for _, discordGuild := range event.Guilds {
		fmt.Printf("Adding kick command...\n")
		_, err := s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, KickCommand)
		if err != nil {
			fmt.Printf("Could not add kick command: %v\n", err)
		}

		fmt.Printf("Adding mute command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, MuteCommand)
		if err != nil {
			fmt.Printf("Could not add mute command: %v\n", err)
		}

		fmt.Printf("Adding unmute command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, UnmuteCommand)
		if err != nil {
			fmt.Printf("Could not add unmute command: %v\n", err)
		}

		fmt.Printf("Adding ban command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, BanCommand)
		if err != nil {
			fmt.Printf("Could not add ban command: %v\n", err)
		}

		fmt.Printf("Adding unban command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, UnbanCommand)
		if err != nil {
			fmt.Printf("Could not add unban command: %v\n", err)
		}

		fmt.Printf("Adding removenickname command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, RemoveNicknameCommand)
		if err != nil {
			fmt.Printf("Could not add removenickname command: %v\n", err)
		}

		fmt.Printf("Adding setnickname command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, SetNicknameCommand)
		if err != nil {
			fmt.Printf("Could not add setnickname command: %v\n", err)
		}

		fmt.Printf("Adding slowmode command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, SlowmodeCommand)
		if err != nil {
			fmt.Printf("Could not add slowmode command: %v\n", err)
		}

		fmt.Printf("Adding slowmodeoff command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, SlowmodeOffCommand)
		if err != nil {
			fmt.Printf("Could not add slowmodeoff command: %v\n", err)
		}

		fmt.Printf("Adding purge command...\n")
		_, err = s.ApplicationCommandCreate(event.User.ID, discordGuild.ID, PurgeCommand)
		if err != nil {
			fmt.Printf("Could not add purge command: %v\n", err)
		}
	}
}

var KickCommand = &discordgo.ApplicationCommand{
	Name:                     "kick",
	DefaultMemberPermissions: &kickPermission,
	Description:              "Kick the specified user and notify the user why via DMs.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User being kicked",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "Reason for kick",
			Required:    false,
		},
	},
}

var MuteCommand = &discordgo.ApplicationCommand{
	Name:                     "mute",
	DefaultMemberPermissions: &mutePermission,
	Description:              "Mute the specified user.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User being muted",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "duration",
			Description: "Duration in minutes",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "Reason for mute",
			Required:    false,
		},
	},
}

var UnmuteCommand = &discordgo.ApplicationCommand{
	Name:                     "unmute",
	DefaultMemberPermissions: &mutePermission,
	Description:              "Unmute the specified user.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User being muted",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "Reason for mute",
			Required:    false,
		},
	},
}

var BanCommand = &discordgo.ApplicationCommand{
	Name:                     "ban",
	DefaultMemberPermissions: &banPermission,
	Description:              "Ban the specified user.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User being banned",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "Reason for ban",
			Required:    false,
		},
	},
}

var UnbanCommand = &discordgo.ApplicationCommand{
	Name:                     "unban",
	DefaultMemberPermissions: &banPermission,
	Description:              "Unban the specified user.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User being unbanned",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "Reason for unban",
			Required:    false,
		},
	},
}

var RemoveNicknameCommand = &discordgo.ApplicationCommand{
	Name:                     "removenickname",
	DefaultMemberPermissions: &mutePermission,
	Description:              "Remove the nickname of the specified user.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User whose nickname to rename",
			Required:    true,
		},
	},
}

var SetNicknameCommand = &discordgo.ApplicationCommand{
	Name:                     "setnickname",
	DefaultMemberPermissions: &mutePermission,
	Description:              "Change the nickname of the specified user.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User whose nickname to rename",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "nickname",
			Description: "Nickname to rename user to",
			Required:    true,
		},
	},
}

var SlowmodeCommand = &discordgo.ApplicationCommand{
	Name:                     "slowmode",
	DefaultMemberPermissions: &channelPermission,
	Description:              "Add slowmode to current channel.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "duration",
			Description: "Slowmode interval duration in seconds",
			Required:    true,
		},
	},
}

var SlowmodeOffCommand = &discordgo.ApplicationCommand{
	Name:                     "slowmodeoff",
	DefaultMemberPermissions: &channelPermission,
	Description:              "Remove slowmode from current channel.",
}

var PurgeCommand = &discordgo.ApplicationCommand{
	Name:                     "purge",
	DefaultMemberPermissions: &messagePermission,
	Description:              "Delete a number of messages from a channel.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Channel to purge",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "message-number",
			Description: "Number of messages to purge",
			Required:    true,
		},
	},
}

func (d *Discord) InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "kick":
		d.Kick(s, i)
	case "mute":
		d.Mute(s, i)
	case "ban":
		d.Ban(s, i)
	case "unban":
		d.Unban(s, i)
	case "removenickname":
		d.RemoveNickname(s, i)
	case "setnickname":
		d.SetNickname(s, i)
	case "slowmode":
		d.Slowmode(s, i)
	case "slowmodeoff":
		d.SlowmodeOff(s, i)
	case "purge":
		d.Purge(s, i)
	}
}
