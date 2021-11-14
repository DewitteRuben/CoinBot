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
	max6DecimalPlaces := math.Floor(cData.Price*1000000) / 1000000
	formattedAmount := strconv.FormatFloat(max6DecimalPlaces, 'f', -1, 64)
	return fmt.Sprintf("%s %s%s", strings.ToUpper(cData.Symbol), currencySymbol, formattedAmount)
}

func (t *BotTracker) formatPriceChange(cData api.CoinData, config *BotTrackerConfig) string {
	change := cData.DailyChangePercentage
	abbreviation := config.Timespan

	switch config.Timespan {
	case "24h":
		change = cData.DailyChangePercentage
	case "1h":
		change = cData.HourlyChangePercentage
	case "7d":
		change = cData.SevenDailyChangePercentage
	default:
		abbreviation = "24h"
		fmt.Println("an unknown price change timespan was given", config.Timespan, "defaulting to 24h")
	}

	return fmt.Sprintf("%s: %s%%", abbreviation, strconv.FormatFloat(change, 'f', 2, 64))
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
			fmt.Printf("Getting latest price data price for %s...\n", t.coinID)
			coinData, err := t.coinAPI.GetCoinData(t.coinID, t.config.Currency)
			if err != nil {
				fmt.Println(err.Error())
			}

			fmt.Printf("Updating bot status...")
			formattedPrice := t.formatPrice(coinData)
			err = t.discordAPI.UpdateBotNickname(formattedPrice)
			if err != nil {
				fmt.Println("Error occurred while updating bot nickname:", err.Error())
			}

			formattedPriceChange := t.formatPriceChange(coinData, t.config)
			err = t.discordAPI.UpdateBotActivity("online", discordgo.ActivityTypeWatching, formattedPriceChange)
			if err != nil {
				fmt.Println("Error occurred while updating bot nickname:", err.Error())
			}

			// update the image once (5min cooldown)
			if !t.hasUpdatedImage {
				err = t.discordAPI.UpdateBotAvatar(coinData.Image)
				if err != nil {
					fmt.Println("Error occurred while updating bot avatar:", err.Error())
				} else {
					t.hasUpdatedImage = true
				}
			}

			fmt.Printf("Update completed! Sleeping for: %s\n", interval)

			time.Sleep(interval)
		}
	}()

}
