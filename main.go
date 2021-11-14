package main

import (
	"btcbot/api"
	"btcbot/tracker"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token               string
	Coin                string
	Currency            string
	PriceChangeTimespan string
)

func init() {
	flag.StringVar(&Token, "token", "OTA5MTg2MDU3NDQ1MjEyMTYw.YZAnfw.eDY8THNcwVdmgPGCiUWGWmK2OZI", "Bot Token")
	flag.StringVar(&Coin, "coin", "curve-dao-token", "Coin to track")
	flag.StringVar(&Currency, "currency", "usd", "Currency to represent coin in")
	flag.StringVar(&PriceChangeTimespan, "pChange", "24h", "The price change timespan to look at")
	flag.Parse()
}

func main() {
	session, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	err = session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	config := tracker.BotTrackerConfig{
		CoinID:   Coin,
		Backend:  api.CoinGecko,
		Timespan: PriceChangeTimespan,
		Currency: Currency,
	}

	botTracker, err := tracker.NewBotTracker(config, session)
	if err != nil {
		panic(err)
	}

	botTracker.Track()

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	session.Close()
}
