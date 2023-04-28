package bot

import (
	"balanceBot/balance"
	"balanceBot/setting"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shopspring/decimal"
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
			if update.Message.Text != "/balance" {
				continue
			}

			assets, err := balance.Service().GetBalance()
			if err != nil {
				logrus.Errorf("can't get balance, err: %s", err.Error())
				continue
			}
			msgText := "红豆子目前持仓情况：\n```\n"
			msgText += fmt.Sprintf("%-6s %14s %12s\n", "Symbol", "Amount", "Value (USDT)")
			for _, asset := range assets {
				amount, _ := asset.Amount.Float64()
				value, _ := asset.Value.Float64()
				msgText += fmt.Sprintf("%-6s %14.8f %12.2f\n", asset.Symbol, amount, value)
			}
			totalValue := decimal.Decimal{}
			for _, asset := range assets {
				totalValue = totalValue.Add(asset.Value)
			}
			totalFloat, _ := totalValue.Float64()
			msgText += "\nTotal: " + fmt.Sprintf("%.2f USDT", totalFloat)
			cny2usd, err := balance.GetCnyCurrency("USD")
			if err != nil {
				logrus.Errorf("can't get cny currency, err: %s", err.Error())
				continue
			}
			msgText += "\nTotal: " + fmt.Sprintf("%.2f CNY", totalFloat/cny2usd)
			msgText += "```"

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			msg.ParseMode = "MarkdownV2"
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
