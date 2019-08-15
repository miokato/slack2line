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

func createBot() *linebot.Client {
	secret := os.Getenv("LINE_SECRET")
	token := os.Getenv("LINE_TOKEN")
	bot, err := linebot.New(
		secret, token,
	)
	if err != nil {
		log.Fatal(err)
	}
	return bot
}

type slackMessageFormat struct {
	Text      string `json:"text"`
	Username  string `json: "username"`
	IconEmoji string `json: "icon_emoji"`
	IconURL   string `json: "icon_url"`
	Channel   string `json: "channel"`
}

func createSlackMessage(rawMessage string) string {
	format := slackMessageFormat{
		Text:      rawMessage,
		Username:  "non",
		IconEmoji: ":gopher:",
		IconURL:   "",
		Channel:   "",
	}

	json, err := json.Marshal(format)
	if err != nil {
		log.Fatal(err)
	}
	return string(json)
}

func postToSlack(mes string) error {
	// token := os.Getenv("SLACK_TOKEN")
	// outtoken := os.Getenv("SLACK_OUTGOING_TOKEN")
	payload := createSlackMessage(mes)

	u := os.Getenv("SLACK_URL")
	resp, err := http.PostForm(u, url.Values{"payload": {payload}})
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return err
}

func postToLine(bot *linebot.Client, mes string) {
	uid := os.Getenv("LINE_USER_ID")
	_, err := bot.PushMessage(uid, linebot.NewTextMessage(mes)).Do()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
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
		postToLine(bot, raw)
	})

	// line to slack
	r.POST("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				log.Fatal(err)
			}
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
