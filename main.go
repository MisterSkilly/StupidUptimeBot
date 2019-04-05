package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"os"
	"time"
	"io/ioutil"
)

type Configuration struct {
	Port            int
	Password        string
	BotToken        string
	AlertUser       string
	AlertChatID     int64
	Minutes         int
	AdminUserID     int
	HetznerUser     string
	HetznerPassword string
	HetznerIP       string
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
				msg := tgbotapi.NewMessage(configuration.AlertChatID, "You got hacked, son! "+configuration.AlertUser+"\nFeel free to say 'restart please'.")
				bot.Send(msg)
			}
		}
	}()

	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates, err := bot.GetUpdatesChan(u)
		if err != nil {
			panic(err)
		}

		for update := range updates {
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}
			if update.Message.From.ID == configuration.AdminUserID && update.Message.Text == "restart please" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Alright buddy, restarting for you!")
				bot.Send(msg)

				restartHetzner()

				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Restarted.")
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

func restartHetzner(){
	rUrl := "https://robot-ws.your-server.de/reset/"+configuration.HetznerIP+"?type=hw"
	fmt.Println("URL:>", rUrl)

	req, err := http.NewRequest("POST", rUrl, nil)
	req.SetBasicAuth(configuration.HetznerUser, configuration.HetznerPassword)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
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
