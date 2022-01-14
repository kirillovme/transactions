package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"

	"github.com/gocarina/gocsv"
)

const bitSize = 64

type Transaction struct {
	// a UUID of transaction
	ID string `csv:"id"`
	// in USD, typically a value between "0.01" and "1000" USD.
	Amount string `csv:"amount"`
	// a 2-letter country code of where the bank is located
	BankCountryCode string `csv:"bank_country_code"`
}

// Reading everything from csv file transactions
func readAllTransactions(filename string) []*Transaction {
	in, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		fmt.Println("Error insde reading everything from csv file: ", err)
		os.Exit(1)
	}

	defer in.Close()

	transactions := []*Transaction{}

	if err := gocsv.UnmarshalFile(in, &transactions); err != nil {
		panic(err)
	}

	return transactions

}

// Reading json from api_latencies into an interface
func getJsonOnject(filename string) map[string]interface{} {
	jsonFile, err := os.Open(filename)

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}

	json.Unmarshal([]byte(byteValue), &result)

	return result

}

func prioritize(tx []*Transaction, dur int, apiLatencies map[string]interface{}) ([]*Transaction, error) {
	sort.SliceStable(tx, func(i, j int) bool {
		valueFirst, errFirst := strconv.ParseFloat(tx[i].Amount, bitSize)
		valueSecond, errSecond := strconv.ParseFloat(tx[j].Amount, bitSize)

		if errFirst != nil || errSecond != nil {
			fmt.Println("Error with the first value inside sorting in prioritizing: ", errFirst)
			fmt.Println("Error with the second value inside sorting in prioritizing: ", errSecond)
			os.Exit(1)
		}

		return valueFirst/apiLatencies[tx[i].BankCountryCode].(float64) > valueSecond/apiLatencies[tx[j].BankCountryCode].(float64)
	})

	tempTime := 0
	index := 0

	for tempTime < dur {
		transactionDuration := int(apiLatencies[tx[index].BankCountryCode].(float64))

		tempTime += transactionDuration

		if tempTime > dur {
			tempTime = tempTime - transactionDuration
			break
		}

		index++
	}

	fmt.Printf("Total time of the transaction is %vms", tempTime)
	fmt.Println()

	amountTransaction := tx[0:index]

	amountProperties := []float64{}

	for _, transaction := range amountTransaction {
		value, err := strconv.ParseFloat(transaction.Amount, bitSize)
		if err != nil {
			fmt.Println("Error inside prioritize", err)
			os.Exit(1)
		}
		amountProperties = append(amountProperties, value)

	}

	maxTransactionsValue := .0

	for _, amount := range amountProperties {
		maxTransactionsValue += amount
	}

	fmt.Printf("Maximum amount of transactions is %v$ in %vms", maxTransactionsValue, dur)

	return amountTransaction, nil
}

func main() {

	transactions := readAllTransactions("transactions.csv")

	prioritize(transactions, 50, getJsonOnject("api_latencies.json"))

}
