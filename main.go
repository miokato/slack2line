package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func HttpPost(mes string, slack_url string) error {
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
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(body))
	return err
}

func main() {
	line_secret := os.Getenv("LINE_SECRET")
	line_token := os.Getenv("LINE_TOKEN")
	line_user_id := os.Getenv("LINE_USER_ID")
	slack_token := os.Getenv("SLACK_TOKEN")
	slack_outgoing_token := os.Getenv("SLACK_OUTGOING_TOKEN")
	slack_url := os.Getenv("SLACK_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Println(slack_token, slack_outgoing_token)

	bot, err := linebot.New(
		line_secret, line_token,
	)
	if err != nil {
		log.Fatal(err)
	}

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
					fmt.Println(message.Text)
					HttpPost(message.Text, slack_url)
				}
			}
		}
	})
	r.Run(":" + port)
}
