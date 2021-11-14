package discord

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type DiscordAPI struct {
	session *discordgo.Session
}

func NewDiscordAPI(session *discordgo.Session) DiscordAPI {
	return DiscordAPI{session: session}
}

func (dc *DiscordAPI) getAllGuildIDs() ([]string, error) {
	guilds, err := dc.session.UserGuilds(100, "", "")
	if err != nil {
		return nil, err
	}

	var ids = []string{}
	for _, guild := range guilds {
		ids = append(ids, guild.ID)
	}

	return ids, nil
}

func (dc *DiscordAPI) UpdateBotAvatar(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	contentType := http.DetectContentType(img)
	base64img := base64.StdEncoding.EncodeToString(img)

	avatar := fmt.Sprintf("data:%s;base64,%s", contentType, base64img)
	_, err = dc.session.UserUpdate("", "", "", avatar, "")
	if err != nil {
		return err
	}

	return nil
}

func (dc *DiscordAPI) UpdateBotNickname(value string) error {
	guildIDs, err := dc.getAllGuildIDs()
	if err != nil {
		return err
	}

	for _, id := range guildIDs {
		dc.session.GuildMemberNickname(id, "@me", value)
	}

	return nil
}

func (dc *DiscordAPI) UpdateBotActivity(status string, activityType discordgo.ActivityType, text string) error {
	return dc.session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status: status,
		Activities: []*discordgo.Activity{
			{
				Type: activityType,
				Name: text,
			},
		},
	})
}
