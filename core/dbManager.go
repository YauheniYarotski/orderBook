package core

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/KristinaEtc/slflog"
	_ "github.com/lib/pq"
)

type DbExchange struct {
	name    string
	//Tickers []DbTicker
}

//type DbTicker struct {
//	TargetCurrency    currencies.Currency
//	ReferenceCurrency currencies.Currency
//	Rate              float64
//	TimpeStamp        time.Time
//	isCalculated 		bool
//}

type DbManager struct {
	db *sql.DB
}

type DbRate struct {
	exchangeTitle string
	targetCode    string
	referenceCode string
	timeStamp     time.Time
	rate          float64
}

func NewDbManager(configuration DBConfiguration) *DbManager {
	manager := DbManager{}
	manager.db = manager.connectDb(configuration)
	return &manager
}

func (b *DbManager) connectDb(configuration DBConfiguration) *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		configuration.User, configuration.Password, configuration.Name)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Errorf("connectDb:DbManager:sql.Open %v", err.Error())
	} else {
		log.Infof("Db connected")
	}
	return db
	//defer db.Close()
}

func (b *DbManager) FillDb(exchangeBooks []ExchangeBook) {
	//fmt.Println(exchangeBook)
	for _, exchangeBook := range exchangeBooks {
		for _, coinBook := range exchangeBook.Coins {
			//for price, amount := range coinBook.PriceLevels.Bids.Range()
			coinBook.PriceLevels.Bids.Range(func(key, value interface{}) bool {
				b.insertSaBook(exchangeBook.Exchange.String(), coinBook.Pair.TargetCurrency, coinBook.Pair.ReferenceCurrency, key.(string), false, 0, value.(string))
				return true
			})

			//for price, amount := range coinBook.PriceLevels.Asks
			coinBook.PriceLevels.Asks.Range(func(key, value interface{}) bool {
				b.insertSaBook(exchangeBook.Exchange.String(), coinBook.Pair.TargetCurrency, coinBook.Pair.ReferenceCurrency, key.(string), true, 0, value.(string))
				return true
			})
		}
	}
	b.fillBookFromSA()
}

//
//func (b *DbManager) insert(exchange *DbExchange) {
//	//fmt.Println("# Inserting values")
//
//	_, err := b.db.Exec("INSERT INTO exchanges(title,create_date) VALUES($1,$2) ON CONFLICT DO NOTHING;", exchange.name, time.Now())
//	//rows.Close()
//	checkErr(err)
//	//b.db.
//	//fmt.Println("inserted rows:", rows)
//}
//
//func (b *DbManager) insertCurrency(currency currencies.Currency) {
//	//fmt.Println("# Inserting values")
//
//	_, err := b.db.Exec("INSERT INTO currencies(code, title, create_date, native_id) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING;", currency.CurrencyCode(), currency.CurrencyName(), time.Now(), currency)
//	//rows.Close()
//	checkErr(err)
//	//b.db.
//	//fmt.Println("inserted rows:", rows)
//}
//
func (b *DbManager) insertSaBook(exchange_title string, target_currency Currency, reference_currency Currency, priceString string, isAsk bool, count int, amountString string) {
	price, _ := strconv.ParseFloat(priceString, 64)
	amount, _ := strconv.ParseFloat(amountString, 64)
	_, err := b.db.Exec("INSERT INTO sa_book_orders(exchange_title, target_title, target_code, target_native_id, reference_title, reference_code, reference_native_id, price, is_ask, count, amount, time_stamp) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);", exchange_title, target_currency.CurrencyName(), target_currency.CurrencyCode(),target_currency, reference_currency.CurrencyName(), reference_currency.CurrencyCode(),reference_currency, price, isAsk, count, amount, time.Now())
	if err != nil {
		log.Errorf("DbManager:insertSaRate:b.db.Exec %v", err.Error())
	}
	//b.db.
}

func (b *DbManager) fillBookFromSA() {
	_, err := b.db.Exec("SELECT fill_book()")
	if err != nil {
		log.Errorf("DbManager:fillRateFromSA:b.db.Exec %v", err.Error())
	}
}

func (b *DbManager) getRates(timeStamp time.Time, exchangeTitle string, targetCode string, refereciesCodes []string) []DbRate {
	var s = StringSlice{}
	s = refereciesCodes
	rows, err := b.db.Query("SELECT * from getRates($1, $2, $3, $4)", timeStamp, exchangeTitle, targetCode, s)
	if err != nil {
		log.Errorf("DbManager:getRates:b.db.Query %v", err.Error())
	}

	var dbRates = []DbRate{}

	for rows.Next() {
		dbRate := DbRate{}
		var exchange_title string
		var target_code string
		var reference_code string
		var time_stamp time.Time
		var rate float64
		err = rows.Scan(&exchange_title, &target_code, &reference_code, &time_stamp, &rate)
		if err != nil {
			log.Errorf("DbManager:getRates:rows.Scan %v", err.Error())
		}
		dbRate.exchangeTitle = exchange_title
		dbRate.targetCode = target_code
		dbRate.referenceCode = reference_code
		dbRate.timeStamp = time_stamp
		dbRate.rate = rate
		//fmt.Println("exchange_title | target_code | reference_code | time_stamp | rate")
		//fmt.Println(exchange_title, target_code, reference_code, time_stamp, rate)

		dbRates = append(dbRates, dbRate)
	}
	rows.Close()
	return dbRates
}

type StringSlice []string

func (stringSlice StringSlice) Value() (driver.Value, error) {
	var quotedStrings []string
	for _, str := range stringSlice {
		quotedStrings = append(quotedStrings, strconv.Quote(str))
	}
	value := fmt.Sprintf("{ %s }", strings.Join(quotedStrings, ","))
	return value, nil
}

