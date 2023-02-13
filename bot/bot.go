package bot

import (
	"balanceBot/balance"
	"balanceBot/setting"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type ServiceObj struct {
}

var service *ServiceObj

func Service() *ServiceObj {
	if service == nil {
		service = &ServiceObj{}
		go Run()
	}
	return service
}

func Run() {
	bot, err := tgbotapi.NewBotAPI(setting.Config().BotKey)
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	logrus.Debugf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			logrus.Debugf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			assets, err := balance.Service().GetBalance()
			if err != nil {
				logrus.Errorf("can't get balance, err: %s", err.Error())
				continue
			}
			msgText := ""
			for _, asset := range assets {
				msgText += fmt.Sprintf("%s %s\n", asset.Symbol, asset.Amount.String())
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
