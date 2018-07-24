package api

import (
	//"fmt"
	//"strconv"
	//"encoding/json"

	"net/http"
	"time"
	//"fmt"
	"fmt"
	"io/ioutil"
	"sync"
	"errors"

)

type RestApi struct {
	httpClient *http.Client
}

type RestApiReposponse struct {
	Message *[]byte
	Err     *error
}


func NewRestApi() *RestApi {
	var api = RestApi{}
	api.httpClient = &http.Client{Timeout: time.Second * 10}
	return &api
}

var (
	//Poloniex says we are allowed 6 req/s
	//but this is not true if you don't want to see
	//'nonce must be greater than' error 3 req/s is the best option.
	throttle = time.Tick(time.Second / 3)
)

var (
	ConnectError    = "[ERROR] Connection could not be established!"
	RequestError    = "[ERROR] NewRequest Error!"
	SetApiError     = "[ERROR] Set the API KEY and API SECRET!"
	PeriodError     = "[ERROR] Invalid Period!"
	TimePeriodError = "[ERROR] Time Period incompatibility!"
	TimeError       = "[ERROR] Invalid Time!"
	StartTimeError  = "[ERROR] Start Time Format Error!"
	EndTimeError    = "[ERROR] End Time Format Error!"
	LimitError      = "[ERROR] Limit Format Error!"
	ChannelError    = "[ERROR] Unknown Channel Name: %s"
	SubscribeError  = "[ERROR] Already Subscribed!"
	WSTickerError   = "[ERROR] WSTicker Parsing %s"
	OrderBookError  = "[ERROR] MarketUpdate OrderBook Parsing %s"
	NewTradeError   = "[ERROR] MarketUpdate NewTrade Parsing %s"
	ServerError     = "[SERVER ERROR] Response: %s"
)

func Error(msg string, args ...interface{}) error {
	if len(args) > 0 {
		return errors.New(fmt.Sprintf(msg, args))
	} else {
		return errors.New(msg)
	}
}

type Logger struct {
	isOpen bool
	Lock   *sync.Mutex
}


func (p *RestApi) PublicRequest(urlString string, responseCh chan <- RestApiReposponse) {

	<-throttle

	//TODO - check if close is needed
	//defer close(responseCh)
	//defer close(errorCh)


	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		fmt.Println("error creating request:", err)
		//errorCh <- Error(RequestError)
		//return
	}

	req.Header.Add("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		fmt.Println("error sending request:", err)
		//errorCh <- Error(ConnectError)
		//return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading response:", err)
		//errorCh <- err
		//return
	}

	restApiResponse := RestApiReposponse{&body, &err}

	responseCh <- restApiResponse
	//errorCh <- nil
}
