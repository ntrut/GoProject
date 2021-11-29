package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/jamespearly/loggly"
)

type cryptoData struct {
	Key struct {
		Id                string `json:"id"`
		Rank              string `json:"rank"`
		Symbol            string `json:"symbol"`
		Name              string `json:"name"`
		Supply            string `json:"supply"`
		MaxSupply         string `json:"maxSupply"`
		MarketCapUsd      string `json:"marketCapUsd"`
		VolumeUsd24Hr     string `json:"volumeUsd24Hr"`
		PriceUsd          string `json:"priceUsd"`
		ChangePercent24hr string `json:"changePercent24hr"`
		Vwap24Hr          string `json:"vwap24Hr"`
		Explorer          string `json:"explorer"`
	} `json:"data"`
	Timestamp int64 `json:"timestamp"`
}

type Item struct {
	Timestamp         int64  `json:"timestamp"`
	Id                string `json:"id"`
	Rank              string `json:"rank"`
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Supply            string `json:"supply"`
	MaxSupply         string `json:"maxSupply"`
	MarketCapUsd      string `json:"marketCapUsd"`
	VolumeUsd24Hr     string `json:"volumeUsd24Hr"`
	PriceUsd          string `json:"priceUsd"`
	ChangePercent24hr string `json:"changePercent24hr"`
	Vwap24Hr          string `json:"vwap24Hr"`
	Explorer          string `json:"explorer"`
}

func goCode(client *loggly.ClientType, tk *time.Ticker, db *dynamodb.DynamoDB, coin string) {
	//define struct
	var dataStruct cryptoData
	resp, err := http.Get("https://api.coincap.io/v2/assets/" + coin)
	var statusCode int = resp.StatusCode
	if err != nil {
		_ = client.Send("error", "Error using the Get request")
		return
	} else {
		if resp.StatusCode != http.StatusOK {
			_ = client.Send("error", "Error code "+resp.Status)
		} else {
			_ = client.Send("info", "200 Status Code, Success")
		}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		_ = client.Send("error", "Error reading from the body")
	}

	//unmarshal the data
	err = json.Unmarshal(body, &dataStruct)
	if err != nil {
		_ = client.Send("error", "Error on unmarshalling from the body")
	}

	//Print out test data
	fmt.Println("************************************************")
	fmt.Println("Struct Display: ", dataStruct)
	fmt.Println("")
	fmt.Println("-----------------------------------------------")
	fmt.Println("ID: ", dataStruct.Key.Id)
	fmt.Println("Current Price: ", dataStruct.Key.PriceUsd)
	fmt.Println("************************************************")

	//Loggly stuff when everything is successful
	//only do this when we receive status code 200
	if statusCode == http.StatusOK {
		_ = client.Send("info", "Successfully got the request")

		bodysize := strconv.Itoa(len(body))

		_ = client.Send("info", "Amount of data is: "+bodysize)

		/*create a new item and sent this item to dynamodb*/
		//date := ime.Now().Format("01-02-2006")
		item := Item{
			Timestamp:         dataStruct.Timestamp,
			Id:                dataStruct.Key.Id,
			Rank:              dataStruct.Key.Rank,
			Symbol:            dataStruct.Key.Symbol,
			Name:              dataStruct.Key.Name,
			Supply:            dataStruct.Key.Supply,
			MaxSupply:         dataStruct.Key.MaxSupply,
			MarketCapUsd:      dataStruct.Key.MarketCapUsd,
			VolumeUsd24Hr:     dataStruct.Key.VolumeUsd24Hr,
			PriceUsd:          dataStruct.Key.PriceUsd,
			ChangePercent24hr: dataStruct.Key.ChangePercent24hr,
			Vwap24Hr:          dataStruct.Key.Vwap24Hr,
		}

		//marshall map the struct
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			_ = client.Send("error", "Got error marshalling new movie item")
		}
		tableName := "ntrut-Crypto"
		//fmt.Println(av)
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}

		_, err = db.PutItem(input)
		if err != nil {
			_ = client.Send("error", "Got error calling PutItem")
			fmt.Println(err)
			_ = client.Send("info", "Successfully created a db item")
		}

	}
}

func main() {

	// Create an AWS session for US East 1. REE
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	// Create a DynamoDB instance
	db := dynamodb.New(sess)

	//define a client for loggly
	client := loggly.New("nazartrut")

	//create poll 3600
	duration := time.Duration(60) * time.Second //every 1 hour

	tk := time.NewTicker(duration)
	top20crypto := [20]string{"bitcoin", "ethereum", "ripple", "bitcoin-cash", "eos", "stellar", "litecoin", "cardano", "tether", "iota", "tron", "ethereum-classic", "monero", "neo", "dash", "binance-coin", "nem", "tezos", "zcash", "dogecoin"}

	for range tk.C {
		for index := range top20crypto {
			goCode(client, tk, db, top20crypto[index])
		}
	}
}
