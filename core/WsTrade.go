package core


type WsTrade struct {
	Exchange	  string 	`json:"e"`
	Symbol        string 	`json:"s"`
	Price         float64	`json:"p"`
	Quantity      float64 	`json:"q"`
	TradeTime     int64  	`json:"t"`
	IfBid  		bool   		`json:"m"`
}