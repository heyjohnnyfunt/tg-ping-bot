package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sparrc/go-ping"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	TgKey string `json:"tg-bot-key"`
}

func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

// Returns average ping time duration
func Ping(addr string) string {

	pinger, err := ping.NewPinger(addr)
	pinger.SetPrivileged(true)
	if err != nil {
		panic(err)
	}
	pinger.Count = 3
	pinger.Run()                 // blocks until finished
	stats := pinger.Statistics() // get send/receive/rtt stats
	fmt.Println(stats)

	return ((stats.MinRtt + stats.MaxRtt) / 2).String()
}

func main() {

	reg, err := regexp.Compile("^https?://")
	if err != nil {
		log.Fatal(err)
	}

	config := LoadConfiguration("./config.json")
	bot, err := tgbotapi.NewBotAPI(config.TgKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		var answer string

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		words := strings.Fields(update.Message.Text)

		if words[0] == "ping" {

			if len(words) > 1 && len(words[1]) > 0 {

				addr := words[1]
				log.Printf("%s", addr)

				if !reg.MatchString(addr) {
					addr = "http://" + addr
				}

				// check if IP or URL is valid
				_, err1 := url.ParseRequestURI(addr)
				err2 := net.ParseIP(addr) // returns nil if invalid

				if err1 != nil && err2 == nil {
					answer = "Give me valid IP or URL, SOAB!"
				} else {
					addr = reg.ReplaceAllString(addr, "")
					pingTime := Ping(addr)
					answer = fmt.Sprintf("time:%s\npong at:\n%s", pingTime, time.Now().UTC().String())
				}
			} else {
				answer = fmt.Sprintf("pong at:\n%s", time.Now().UTC().String())
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, answer)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
