package core


type CoinBook struct {
	Symbol string 	`json:"symbol"`
	Pair CurrencyPair  	`json:"-"`
	Asks map[string]string		`json:"asks"`
	Bids map[string]string		`json:"bids"`
}


func NewCoinBook(pair CurrencyPair) CoinBook  {
	coinBook := CoinBook{}
	coinBook.Pair = pair
	coinBook.Asks = map[string]string{}
	coinBook.Bids = map[string]string{}
	return coinBook
}
