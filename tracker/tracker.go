package tracker

import (
	"btcbot/api"
	"btcbot/discord"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type BotTracker struct {
	coinID          string
	hasUpdatedImage bool
	coinAPI         api.CoinAPI
	config          *BotTrackerConfig
	discordAPI      *discord.DiscordAPI
}

type BotTrackerConfig struct {
	CoinID   string
	Currency string
	Timespan string
	Backend  api.CoinAPIImplementation
}

func NewBotTracker(config BotTrackerConfig, session *discordgo.Session) (BotTracker, error) {
	var impl api.CoinAPI
	if config.Backend == api.CoinGecko {
		impl = api.NewCoinGeckoAPI()
	}

	if config.CoinID == "" {
		return BotTracker{}, errors.New("coin id is missing in config")
	}

	if config.Currency == "" {
		return BotTracker{}, errors.New("currency is missing in config")
	}

	if config.Timespan == "" {
		return BotTracker{}, errors.New("timespan is missing in config")
	}

	if config.Backend == 0 {
		return BotTracker{}, errors.New("backend is missing in config")
	}

	discordAPI := discord.NewDiscordAPI(session)

	return BotTracker{
		coinID:          config.CoinID,
		coinAPI:         impl,
		config:          &config,
		discordAPI:      &discordAPI,
		hasUpdatedImage: false,
	}, nil
}

func (t *BotTracker) formatPrice(cData api.CoinData) string {
	// TODO: dynamic currency symbol
	currencySymbol := "$"
	pChange, _ := t.getPriceChangePercentage(cData)
	max6DecimalPlaces := math.Floor(cData.Price*1000000) / 1000000
	formattedAmount := strconv.FormatFloat(max6DecimalPlaces, 'f', -1, 64)
	var changeDirectionSymbol string

	// stonks or no stonks
	if pChange >= 0 {
		changeDirectionSymbol = "📈"
	} else {
		changeDirectionSymbol = "📉"
	}

	return fmt.Sprintf("%s %s%s %s", strings.ToUpper(cData.Symbol), currencySymbol, formattedAmount, changeDirectionSymbol)
}

func (t *BotTracker) getPriceChangePercentage(cData api.CoinData) (float64, string) {
	pChange := cData.DailyChangePercentage
	abbr := t.config.Timespan

	switch t.config.Timespan {
	case "24h":
		pChange = cData.DailyChangePercentage
	case "1h":
		pChange = cData.HourlyChangePercentage
	case "7d":
		pChange = cData.SevenDailyChangePercentage
	default:
		abbr = "24h"
		log.Warn("an unknown price change timespan was given", t.config.Timespan, "defaulting to 24h")
	}

	return pChange, abbr
}

func (t *BotTracker) formatPriceChange(cData api.CoinData) string {
	pChange, abbr := t.getPriceChangePercentage(cData)

	return fmt.Sprintf("%s: %s%%", abbr, strconv.FormatFloat(pChange, 'f', 2, 64))
}

func (t *BotTracker) Track(intervals ...time.Duration) {
	var interval time.Duration
	if len(intervals) == 0 {
		interval = time.Minute * 3
	} else {
		interval = intervals[0]
	}

	go func() {
		for {
			log.Infof("Getting latest price data price for %s...\n", t.coinID)
			coinData, err := t.coinAPI.GetCoinData(t.coinID, t.config.Currency)
			if err != nil {
				log.Error("Error occurred while getting coin data", err)

				time.Sleep(time.Minute) // wait and try again

				continue
			}

			log.Info("Updating bot status...")
			formattedPrice := t.formatPrice(coinData)
			err = t.discordAPI.UpdateBotNickname(formattedPrice)
			if err != nil {
				log.Error("Error occurred while updating bot nickname:", err)
			}

			formattedPriceChange := t.formatPriceChange(coinData)
			err = t.discordAPI.UpdateBotActivity("online", discordgo.ActivityTypeWatching, formattedPriceChange)
			if err != nil {
				log.Error("Error occurred while updating bot nickname:", err)
			}

			// update the image once (5min cooldown)
			if !t.hasUpdatedImage {
				err = t.discordAPI.UpdateBotAvatar(coinData.Image)
				if err != nil {
					log.Error("Error occurred while updating bot avatar:", err)
				} else {
					t.hasUpdatedImage = true
				}
			}

			log.Infof("Update completed! Sleeping for: %s\n", interval)

			time.Sleep(interval)
		}
	}()

}
