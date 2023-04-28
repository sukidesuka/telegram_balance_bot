package balance

import (
	"balanceBot/setting"
	"context"
	"encoding/json"
	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"sort"
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
				Value:  decimal.Zero,
			})

		}
	}

	// get bitcoin balance from blockchain
	for _, address := range setting.Config().BitcoinAddresses {
		balance, err := getBitcoinBalanceFromBlockchain(address)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &Asset{
			Symbol: "BTC",
			Amount: *balance,
			Value:  decimal.Zero,
		})
	}

	// get ethereum balance from blockchain
	for _, address := range setting.Config().EthereumAddresses {
		balance, err := getEthereumBalanceFromBlockchain(address)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &Asset{
			Symbol: "ETH",
			Amount: *balance,
			Value:  decimal.Zero,
		})
	}

	// aggregate assets
	assetsMap := make(map[string]*Asset)
	for _, asset := range assets {
		if _, ok := assetsMap[asset.Symbol]; !ok {
			assetsMap[asset.Symbol] = &Asset{
				Symbol: asset.Symbol,
				Amount: decimal.Zero,
				Value:  decimal.Zero,
			}
		}
		assetsMap[asset.Symbol].Amount = assetsMap[asset.Symbol].Amount.Add(asset.Amount)
	}
	assets = make([]*Asset, 0)
	for _, asset := range assetsMap {
		assets = append(assets, asset)
	}

	// calc market value
	prices, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	priceMap := make(map[string]decimal.Decimal)
	for _, price := range prices {
		priceMap[price.Symbol] = decimal.RequireFromString(price.Price)
	}
	for _, asset := range assets {
		if asset.Symbol == "USDT" {
			asset.Value = asset.Amount
			continue
		}

		stableDollars := []string{"USDT", "BUSD", "TUSD"}
		var price decimal.Decimal
		for _, stableDollar := range stableDollars {
			if p, ok := priceMap[asset.Symbol+stableDollar]; ok {
				price = p
				break
			}
		}

		asset.Value = asset.Amount.Mul(price)
	}

	// sort assets
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Value.Cmp(assets[j].Value) > 0
	})

	return assets, nil
}

type BtcAddressDetail struct {
	Base58 struct {
		Hash    string `json:"hash"`
		Version int    `json:"version"`
	} `json:"base58"`
	Encoding        string `json:"encoding"`
	ValidateAddress struct {
		IsValid      bool   `json:"isvalid"`
		Address      string `json:"address"`
		ScriptPubKey string `json:"scriptPubKey"`
		IsScript     bool   `json:"isscript"`
		IsWitness    bool   `json:"iswitness"`
	} `json:"validateaddress"`
	ElectrumScripthash string `json:"electrumScripthash"`
	TxHistory          struct {
		TxCount            int            `json:"txCount"`
		Txids              []string       `json:"txids"`
		BlockHeightsByTxid map[string]int `json:"blockHeightsByTxid"`
		BalanceSat         int            `json:"balanceSat"`
		Request            struct {
			Limit  int    `json:"limit"`
			Offset int    `json:"offset"`
			Sort   string `json:"sort"`
		} `json:"request"`
	} `json:"txHistory"`
}

func getBitcoinBalanceFromBlockchain(address string) (*decimal.Decimal, error) {
	resp, err := http.Get("https://bitcoinexplorer.org/api/address/" + address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	detail := &BtcAddressDetail{}
	err = json.Unmarshal(bodyBytes, detail)
	if err != nil {
		return nil, err
	}

	balance := decimal.NewFromInt(int64(detail.TxHistory.BalanceSat))
	balance = balance.Div(decimal.NewFromFloat(100000000))

	return &balance, nil
}

type EthAddressDetail struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

func getEthereumBalanceFromBlockchain(address string) (*decimal.Decimal, error) {
	apiURL, err := url.Parse("https://api.etherscan.io/api")
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("module", "account")
	values.Set("action", "balance")
	values.Set("address", address)
	values.Set("tag", "latest")
	values.Set("apikey", setting.Config().EtherScanApiKey)

	apiURL.RawQuery = values.Encode() // 将原有请求参数替换为新构造的请求参数

	response, err := http.Get(apiURL.String())
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("eth balance response: %s", string(body))

	detail := &EthAddressDetail{}
	err = json.Unmarshal(body, detail)
	if err != nil {
		return nil, err
	}

	balance, err := decimal.NewFromString(detail.Result)
	if err != nil {
		return nil, err
	}

	balance = balance.Div(decimal.NewFromFloat(1000000000000000000))

	return &balance, nil
}

type ExchangeRateAPIResponse struct {
	Provider           string             `json:"provider"`
	WarningUpgradeToV6 string             `json:"WARNING_UPGRADE_TO_V6"`
	Terms              string             `json:"terms"`
	Base               string             `json:"base"`
	Date               string             `json:"date"`
	TimeLastUpdated    int                `json:"time_last_updated"`
	Rates              map[string]float64 `json:"rates"`
}

func GetCnyCurrency(dist string) (float64, error) {
	resp, err := http.Get("https://api.exchangerate-api.com/v4/latest/CNY")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var exchangeRateAPIResponse ExchangeRateAPIResponse
	err = json.Unmarshal(bodyBytes, &exchangeRateAPIResponse)

	return exchangeRateAPIResponse.Rates[dist], nil
}
