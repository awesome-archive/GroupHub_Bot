package main
import (
	"log"
	"encoding/base64"
	"strings"
	"net/http"
	"time"
	"os"
	"github.com/tucnak/telebot"
	"github.com/bitly/go-simplejson"
)

var dict map[string][]string = make(map[string][]string)
var bot *telebot.Bot
var Tags []string

func main() {
	token := os.Getenv("TELEBOT_TOKEN")
	if len(token) > 0 {
		log.Printf("Telegram Bot Token: %v\n", token)
	} else {
		log.Fatalln("Please set 'TELEBOT_TOKEN' from environment variable")
	}

	startInline()


	var err error
	bot, err = telebot.NewBot(token)
	if err != nil {
		log.Fatalln(err)
	}
	bot.Messages = make(chan telebot.Message, 1000)
	bot.Queries = make(chan telebot.Query, 1000)
	go messages()
	go queries()
	bot.Start(1*time.Second)
}

func startInline() {
	resp, err := http.Get("https://raw.githubusercontent.com/guo-yu/o3o/master/yan.json")
	if err != nil {
		log.Fatalln(err)
	}
	js, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	i := 0
	for {

		Tag, err := js.Get("list").GetIndex(i).Get("tag").String()
		if err != nil {
			break
		}
		Tags = append(Tags, Tag)
		j := 0
		for {
			Yan, err := js.Get("list").GetIndex(i).Get("yan").GetIndex(j).String()
			if err != nil {
				break
			}
			dict[Tag] = append(dict[Tag], Yan)
			j++
		}
		i++
	}


	log.Println("--- Dictionary initialized. ---")
}

func messages() {
	//read groups information

	for message := range bot.Messages {
		resp, err := http.Get("https://raw.githubusercontent.com/livc/GroupHub_Bot/master/groups.json")
		if err != nil {
			log.Fatalln(err)
		}
		js, err := simplejson.NewFromReader(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("--- new message ---")
		log.Printf("Received a message from %s with the text: %s\n", message.Sender.Username, message.Text)
		switch message.Text {
			case "/start":
				sta :=  "GroupHub致力于收录tg中文圈优质群组。\n"+
						"项目地址: https://github.com/livc/GroupHub_Bot\n\n"+
						"Bot: @GroupHub_bot\n"+
						"广播站: @GroupHub\n"+
						"交流群: @GroupHub_Chat\n"+
						"群组收录更新: 将【群组名、分类、链接】发送到https://github.com/livc/GroupHub_Bot\n"+
						"BUG提交/功能建议: https://github.com/livc/GroupHub_Bot/issues\n\n"+
						"/groups 查询群组"
				bot.SendMessage(message.Chat, sta, nil)
			case "/groups":
				bot.SendMessage(message.Chat, "请选择分类：", &telebot.SendOptions{
							ReplyMarkup: telebot.ReplyMarkup{
								ForceReply: true,
								Selective: true,

								CustomKeyboard: [][]string{
									[]string{"ACG", "软件", "科学上网"},
									[]string{"linux", "社区", "Geek"},
									[]string{"编程", "城市", "书影音"},
									[]string{"政治", "资源", "其他"},
								},
							},
						},
				)
			case "ACG", "软件", "科学上网", "linux", "社区", "Geek", "编程", "城市", "书影音", "政治", "资源", "其他":
				deal(message, js)
		}
	}
}

func deal(message telebot.Message, js *simplejson.Json) {
	var all string
	i := 0
	for {
		text, err := js.Get(message.Text).GetIndex(i).Get("TEXT").String()
		if err != nil {
			break	
		}
		decodedT, err := base64.StdEncoding.DecodeString(text)
		text = string(decodedT)
//		text = strings.Replace(text, "_", "\\_", -1)    //solve the confilc of markdown "_"
//		文本中带有下划线的需手动在下划线前面加'\'解析 部分群组邀请链接中也含有下划线导致无法进行替换
		all = all+"\n"+text
		i++
	}
//	log.Println(all)
	bot.SendMessage(message.Chat, all, &telebot.SendOptions{ParseMode:"Markdown"})
}

func queries() {
	for query := range bot.Queries {
		log.Println("---new query---")	
		log.Println("from:", query.From.Username)
		log.Println("text:", query.Text)

		results := make([]telebot.InlineQueryResult, 0, 19)
R:		for _, tag := range Tags {
			if strings.Contains(tag, query.Text) {
				for _, yan := range dict[tag] {
					results = append(results, &telebot.InlineQueryResultArticle{Title: yan, Description: tag, Text: yan})
					if len(results) >= cap(results) {
						break R
					}
				}
			}	
		}

		// Build a response object to answer the query
		response := telebot.QueryResponse{
			Results: results,
			IsPersonal: true,
		}

		// And finally send the response.
		if err := bot.AnswerInlineQuery(&query, &response); err != nil {
			log.Println("Failed to respond to query:", err)
		}
	}
}
