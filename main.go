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

//chan to communicate http and ticker events
var received = make(chan int, 10)

var bot *tgbotapi.BotAPI
var configuration = Configuration{}

func main() {
	var path string
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		path = "conf.json"
	}
	//Loading config
	loadConfig(path)

	//Initializing bot
	var err error
	bot, err = tgbotapi.NewBotAPI(configuration.BotToken)
	if err != nil {
		panic(err)
	}

	//Starting goroutines
	go sendHelp()
	//ticker sending a 0 every X minutes to channel
	ticker := time.NewTicker(time.Duration(configuration.Minutes) * time.Minute)
	go func() {
		for range ticker.C {
			received <- 0
		}
	}()

	//handling http
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", configuration.Port), nil))
}

// sendHelp does the actual check. For every 0 received (tick from ticker) a counter increases by 1. When this counter
// reaches 5, the user is alerted on every subsequent tick or event as something has gone wrong. In a working system,
// the counter should never reach 5 as it is decreased by 1 every time the webserver receives a request with the correct
// password and sends a 1 to the channel. This means that for this uptime bot to work, the "Minutes" value in the config
// must be set to the same value as the to-be-checked system's http-request-sending cron.
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
			msg := tgbotapi.NewMessage(configuration.AlertChatID, "You got hacked, son! "+configuration.AlertUser)
			bot.Send(msg)
		}
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == configuration.Password {
		fmt.Println("correct password")
		w.Write([]byte("reset timer."))
		// sending http event to channel
		received <- 1
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
