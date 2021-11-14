package main

import (
	"btcbot/api"
	"btcbot/tracker"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
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

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	session, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
		return
	}

	err = session.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
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

	log.Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	session.Close()
}
