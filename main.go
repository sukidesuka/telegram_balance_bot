package main

import (
	"balanceBot/bot"
	"balanceBot/setting"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"time"
)

type Asset struct {
	Symbol string
	Amount decimal.Decimal
}

func main() {
	setting.Setup()
	bot.Service()
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)

	time.Sleep(time.Hour)

}
