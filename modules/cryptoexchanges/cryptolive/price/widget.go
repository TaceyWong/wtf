package price

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var baseURL = "https://min-api.cryptocompare.com/data/price"
var ok = true

// Widget define wtf widget to register widget later
type Widget struct {
	*list
	settings *Settings

	Result string

	RefreshInterval int
}

// NewWidget Make new instance of widget
func NewWidget(settings *Settings) *Widget {
	widget := Widget{
		settings: settings,
	}

	widget.setList()

	return &widget
}

func (widget *Widget) setList() {
	widget.list = &list{}

	for symbol, currency := range widget.settings.currencies {
		toList := widget.getToList(symbol)
		widget.list.addItem(symbol, currency.displayName, toList)
	}
}

/* -------------------- Exported Functions -------------------- */

// Refresh & update after interval time
func (widget *Widget) Refresh(wg *sync.WaitGroup) {
	if len(widget.list.items) == 0 {
		return
	}

	widget.updateCurrencies()

	if !ok {
		widget.Result = fmt.Sprint("Please check your internet connection!")
		return
	}
	widget.display()
	wg.Done()
}

/* -------------------- Unexported Functions -------------------- */

func (widget *Widget) display() {
	str := ""

	for _, item := range widget.list.items {
		str += fmt.Sprintf(
			" [%s]%s[%s] (%s)\n",
			widget.settings.colors.from.name,
			item.displayName,
			widget.settings.colors.from.name,
			item.name,
		)
		for _, toItem := range item.to {
			str += fmt.Sprintf(
				"\t[%s]%s: [%s]%f\n",
				widget.settings.colors.to.name,
				toItem.name,
				widget.settings.colors.to.price,
				toItem.price,
			)
		}
		str += "\n"
	}

	widget.Result = fmt.Sprintf("\n%s", str)
}

func (widget *Widget) getToList(symbol string) []*toCurrency {
	var toList []*toCurrency

	for _, to := range widget.settings.currencies[symbol].to {
		toList = append(toList, &toCurrency{
			name:  to.(string),
			price: 0,
		})
	}

	return toList
}

func (widget *Widget) updateCurrencies() {
	defer func() {
		recover()
	}()
	for _, fromCurrency := range widget.list.items {

		var (
			client       http.Client
			jsonResponse cResponse
		)

		client = http.Client{
			Timeout: time.Duration(5 * time.Second),
		}

		request := makeRequest(fromCurrency)
		response, err := client.Do(request)

		if err != nil {
			ok = false
		} else {
			ok = true
		}

		defer response.Body.Close()

		_ = json.NewDecoder(response.Body).Decode(&jsonResponse)

		setPrices(&jsonResponse, fromCurrency)
	}

}

func makeRequest(currency *fromCurrency) *http.Request {
	fsym := currency.name
	tsyms := ""
	for _, to := range currency.to {
		tsyms += fmt.Sprintf("%s,", to.name)
	}

	url := fmt.Sprintf("%s?fsym=%s&tsyms=%s", baseURL, fsym, tsyms)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
	}

	return request
}

func setPrices(response *cResponse, currencry *fromCurrency) {
	for idx, toCurrency := range currencry.to {
		currencry.to[idx].price = (*response)[toCurrency.name]
	}
}