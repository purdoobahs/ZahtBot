package main

import (
	"context"
	"fmt"
	"os"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
	"github.com/purdoobahs/ZahtBot/internal/cache"
	"github.com/purdoobahs/ZahtBot/internal/cache/memory"
	"github.com/sirupsen/logrus"
)

// ZahtBot is the Discord ZahtBot.
type ZahtBot struct {
	*disgord.Client

	voiceStateCache cache.VoiceStateCache

	dca            []byte
	activeChannels map[disgord.Snowflake]interface{}
}

// NewZahtBot creates a new ZahtBot.
func NewZahtBot(botToken string) (*ZahtBot, error) {
	logger := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	// Zaht audio file
	dca, err := loadDCA()
	if err != nil {
		logger.Debug(fmt.Sprintf("Load DCA error: %+v\n", err))
		return nil, err
	}

	zb := &ZahtBot{
		Client: disgord.New(disgord.Config{
			ProjectName: "ZahtBot",
			BotToken:    botToken,
			Logger:      logger,
		}),

		voiceStateCache: memory.NewVoiceStateCache(),

		dca:            dca,
		activeChannels: map[disgord.Snowflake]interface{}{},
	}

	zb.Ready(func() {
		zb.Logger().Info("ZahtBot is online!")
	})

	// filters
	filter, _ := std.NewMsgFilter(context.Background(), zb)
	filter.SetPrefix("!")

	// !zaht
	zb.On(
		disgord.EvtMessageCreate,

		filter.NotByBot,
		filter.HasPrefix,
		filterNonDM,

		filterNonZahtCommands,
		zb.commandZaht,
	)

	// Voice State Update
	zb.On(disgord.EvtVoiceStateUpdate, zb.updateVoiceState)

	return zb, nil
}

// getVoiceChannelID retrieves the voice channel ID of the message poster, if they're in one
func (zb *ZahtBot) getVoiceChannelID(session disgord.Session, evt *disgord.MessageCreate) disgord.Snowflake {
	_, vs := zb.voiceStateCache.GetVoiceState(evt.Message.Author.ID)
	if vs == nil {
		zb.Logger().Info(fmt.Sprintf("%s (%s) is not in a voice channel\n", evt.Message.Author.Username, evt.Message.Author.ID))
		return 0
	}

	return vs.ChannelID
}

func (zb *ZahtBot) isVoiceChannelActive(channelID disgord.Snowflake) bool {
	_, ok := zb.activeChannels[channelID]
	return ok
}

func (zb *ZahtBot) lockVoiceChannel(channelID disgord.Snowflake, soundName string) {
	zb.activeChannels[channelID] = soundName
}

func (zb *ZahtBot) unlockVoiceChannel(channelID disgord.Snowflake) {
	delete(zb.activeChannels, channelID)
}
