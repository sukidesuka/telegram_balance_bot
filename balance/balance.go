package balance

import (
	"balanceBot/setting"
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type BaseService struct {
}

var service *BaseService

func Service() *BaseService {
	if service == nil {
		service = &BaseService{}
	}
	return service
}

func (s *BaseService) GetBalance() ([]*Asset, error) {
	assets := make([]*Asset, 0)

	client := binance.NewClient(setting.Config().ApiKey, setting.Config().SecretKey)
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		panic(err)
	}

	for _, asset := range account.Balances {
		free, err := decimal.NewFromString(asset.Free)
		if err != nil {
			return nil, err
		}
		locked, err := decimal.NewFromString(asset.Locked)
		if err != nil {
			return nil, err
		}
		total := free.Add(locked)

		if totalFloat, _ := total.Float64(); totalFloat != 0 {
			logrus.Debugf("asset %s amount %s", asset.Asset, total.String())
			assets = append(assets, &Asset{
				Symbol: asset.Asset,
				Amount: total,
				Value:  0,
			})

		}
	}

	return assets, nil
}
