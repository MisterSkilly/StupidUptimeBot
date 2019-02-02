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

var received = make(chan int, 10)
var bot *tgbotapi.BotAPI
var configuration = Configuration{}

func main() {
	loadConfig()

	var err error
	bot, err = tgbotapi.NewBotAPI(configuration.BotToken)
	if err != nil {
		panic(err)
	}

	go sendHelp()
	ticker := time.NewTicker(time.Duration(configuration.Minutes) * time.Second)
	go func() {
		for range ticker.C {
			received <- 0
		}
	}()

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", configuration.Port), nil))
}

func sendHelp() {
	tickerCounter := 0
	for r := range received {
		if r == 0 {
			fmt.Println("received ticker")
			tickerCounter++
		} else {
			fmt.Println("received http")
			tickerCounter--
		}

		fmt.Println("tickerCounter:", tickerCounter)

		if tickerCounter > 5 {
			u := tgbotapi.NewUpdate(0)
			u.Timeout = 60

			msg := tgbotapi.NewMessage(configuration.AlertChatID, "You got hacked, son! "+configuration.AlertUser)
			bot.Send(msg)
		}
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == configuration.Password {
		fmt.Println("correct password")
		w.Write([]byte("good job"))
		received <- 1
		return
	}
	w.Write([]byte("tf u doin here"))
}

func loadConfig() {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	file.Close()
}
