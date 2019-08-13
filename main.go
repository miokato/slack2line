package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

type Payload struct {
	Text      string `json:"text"`
	Username  string `json: "username"`
	IconEmoji string `json: "icon_emoji"`
	IconURL   string `json: "icon_url"`
	Channel   string `json: "channel"`
}

func postToSlack(mes string) error {
	// slack_token := os.Getenv("SLACK_TOKEN")
	// slack_outgoing_token := os.Getenv("SLACK_OUTGOING_TOKEN")
	slack_url := os.Getenv("SLACK_URL")

	payload := Payload{
		Text:      mes,
		Username:  "non",
		IconEmoji: ":gopher:",
		IconURL:   "",
		Channel:   "",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.PostForm(slack_url, url.Values{"payload": {string(jsonPayload)}})
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return err
}

func createBot() *linebot.Client {
	line_secret := os.Getenv("LINE_SECRET")
	line_token := os.Getenv("LINE_TOKEN")
	bot, err := linebot.New(
		line_secret, line_token,
	)
	if err != nil {
		log.Fatal(err)
	}
	return bot
}

func main() {
	line_user_id := os.Getenv("LINE_USER_ID")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	bot := createBot()

	r := gin.New()
	r.Use(gin.Logger())

	// slack to line
	r.POST("/send", func(c *gin.Context) {
		raw := c.PostForm("text")
		bot.PushMessage(line_user_id, linebot.NewTextMessage(raw)).Do()
	})

	// line to slack
	r.POST("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				log.Print(err)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					postToSlack(message.Text)
				}
			}
		}
	})
	r.Run(":" + port)
}
