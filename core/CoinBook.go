package core


type CoinBook struct {
	Symbol string 	`json:"symbol"`
	Pair CurrencyPair  	`json:"-"`
	Asks map[float64]float64		`json:"asks"`
	Bids map[float64]float64		`json:"bids"`

	TotalAsks string		`json:"total_asks"`
	TotalBids string		`json:"total_bids"`
}

func NewCoinBook(pair CurrencyPair) CoinBook  {
	coinBook := CoinBook{}
	coinBook.Pair = pair
	coinBook.Asks = map[float64]float64{}
	coinBook.Bids = map[float64]float64{}
	return coinBook
}

type WsCoinBook struct {
	Symbol string 	`json:"symbol"`
	Pair CurrencyPair  	`json:"-"`
	Asks [][]float64		`json:"asks"`
	Bids [][]float64		`json:"bids"`

	TotalAsks float64		`json:"total_asks"`
	TotalBids float64		`json:"total_bids"`
}


func NewWsCoinBook(pair CurrencyPair) WsCoinBook  {
	coinBook := WsCoinBook{}
	coinBook.Pair = pair
	//coinBook.Asks = []{}
	//coinBook.Bids = [][]float64{}
	return coinBook
}