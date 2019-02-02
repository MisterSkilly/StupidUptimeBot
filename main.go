package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"os"
	"time"
)

type Configuration struct {
	Port        int
	Password    string
	BotToken    string
	AlertUser   string
	AlertChatID int64
	Minutes     int
}

var bot *tgbotapi.BotAPI
var configuration = Configuration{}

// Keeping track of last update
var lastReceived = time.Now()

func main() {
	var path string
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		path = "conf.json"
	}
	// Loading config
	loadConfig(path)

	// Initializing bot
	var err error
	bot, err = tgbotapi.NewBotAPI(configuration.BotToken)
	if err != nil {
		panic(err)
	}

	//Ticker checking every X minutes if the last update isn't too long ago (too long = X + 1 minute to avoid false-positives). If it is, then the bot alerts the user.
	ticker := time.NewTicker(time.Duration(configuration.Minutes) * time.Minute)
	go func() {
		for range ticker.C {
			if time.Now().After(lastReceived.Add(time.Duration(configuration.Minutes)*time.Minute + 1*time.Minute)) {
				msg := tgbotapi.NewMessage(configuration.AlertChatID, "You got hacked, son! "+configuration.AlertUser)
				bot.Send(msg)
			}
		}
	}()

	// Handling http
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", configuration.Port), nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == configuration.Password {
		fmt.Println("correct password")
		w.Write([]byte("reset timer."))
		// Resetting last update to now
		lastReceived = time.Now()
		return
	}
	w.Write([]byte("You're not supposed to be here."))
}

func loadConfig(path string) {
	file, _ := os.Open(path)
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	file.Close()
}
