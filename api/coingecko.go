package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type CoinData struct {
	Symbol                     string
	Price                      float64
	DailyChangePercentage      float64
	SevenDailyChangePercentage float64
	HourlyChangePercentage     float64
	Currency                   string
	LastUpdatedAt              string
	Image                      string
}

// 'https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd&include_24hr_change=true'
type CoinAPI interface {
	GetCoinData(id string, currency string) (CoinData, error)
}

type HttpClient struct {
	client http.Client
}

type CoinAPIImplementation int64

const (
	CoinGecko CoinAPIImplementation = 1
)

func NewHttpClient() HttpClient {
	client := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	return HttpClient{client: client}
}

func (hc HttpClient) GetJSONObject(url string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, getErr := hc.client.Do(req)
	if getErr != nil {
		return nil, err
	}

	defer func() {
		if res.Body != nil {
			res.Body.Close()
		} else {
			log.Debugf("body for GetJSONObject is empty (%s)", url)
		}
	}()

	data, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	stringResult := string(data)
	if strings.Contains(stringResult, "error") {
		log.Error("error occurred while parsing JSON", stringResult)
		return nil, errors.New("api error has occurred")
	}

	var result = make(map[string]interface{})
	jsonErr := json.Unmarshal(data, &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return result, nil
}

func (hc HttpClient) GetJSONArray(url string) ([]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, getErr := hc.client.Do(req)
	if getErr != nil {
		return nil, err
	}

	defer func() {
		if res.Body != nil {
			res.Body.Close()
		} else {
			log.Debugf("body for getJSONArray is empty (%s)", url)
		}
	}()

	data, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	stringResult := string(data)
	if strings.Contains(stringResult, "error") {
		log.Error("Error occurred while parsing JSON", stringResult)
		return nil, errors.New("api error has occurred")
	}

	var result interface{}
	jsonErr := json.Unmarshal(data, &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return result.([]interface{}), nil
}

type CoinGeckoAPI struct {
	baseURL            string
	httpRequestHandler HttpClient
}

func NewCoinGeckoAPI() CoinGeckoAPI {
	return CoinGeckoAPI{
		baseURL:            "https://api.coingecko.com/api/v3",
		httpRequestHandler: NewHttpClient(),
	}
}

func (cg CoinGeckoAPI) GetCoinData(id string, currency string) (CoinData, error) {
	if currency == "" {
		currency = "usd"
	}

	endpointURL := "/coins/markets"
	params := fmt.Sprintf("?ids=%s&vs_currency=%s&order=market_cap_desc&per_page=100&page=1&sparkline=false&price_change_percentage=1h%%2C24h%%2C7d", id, currency)
	result, err := cg.httpRequestHandler.GetJSONArray(cg.baseURL + endpointURL + params)
	if err != nil {
		return CoinData{}, err
	}

	if len(result) <= 0 {
		return CoinData{}, errors.New("coin data returned no results, check the coin id")
	}

	// result should always return only one results in this situation
	priceObject, ok := result[0].(map[string]interface{})
	if !ok {
		return CoinData{}, errors.New("failed to convert price object to map string interface")
	}

	price, ok := priceObject["current_price"].(float64)
	if !ok {
		return CoinData{}, errors.New("failed to convert price to float64")
	}

	dailyChange, ok := priceObject["price_change_percentage_24h_in_currency"].(float64)
	if !ok {
		return CoinData{}, errors.New("failed to convert 24h change to float64")
	}

	hourlyChange, ok := priceObject["price_change_percentage_1h_in_currency"].(float64)
	if !ok {
		return CoinData{}, errors.New("failed to convert 24h change to float64")
	}

	sevenDailyChange, ok := priceObject["price_change_percentage_7d_in_currency"].(float64)
	if !ok {
		return CoinData{}, errors.New("failed to convert 24h change to float64")
	}

	lastUpdatedAt, ok := priceObject["last_updated"].(string)
	if !ok {
		return CoinData{}, errors.New("failed to convert last updated at to string")
	}

	image, ok := priceObject["image"].(string)
	if !ok {
		return CoinData{}, errors.New("failed to convert image to string")
	}

	symbol, ok := priceObject["symbol"].(string)
	if !ok {
		return CoinData{}, errors.New("failed to convert symbol to string")
	}

	return CoinData{
		Symbol:                     symbol,
		Price:                      price,
		Currency:                   currency,
		DailyChangePercentage:      dailyChange,
		SevenDailyChangePercentage: sevenDailyChange,
		HourlyChangePercentage:     hourlyChange,
		LastUpdatedAt:              lastUpdatedAt,
		Image:                      image,
	}, nil
}
