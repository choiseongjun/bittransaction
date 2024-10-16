package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

type PrevOut struct {
	Type    int    `json:"type"`
	Spent   bool   `json:"spent"`
	Value   int    `json:"value"`
	TxIndex int64  `json:"tx_index"`
	N       int    `json:"n"`
	Script  string `json:"script"`
	Addr    string `json:"addr"`
}

type Input struct {
	Sequence int64   `json:"sequence"`
	Witness  string  `json:"witness"`
	Script   string  `json:"script"`
	Index    int     `json:"index"`
	PrevOut  PrevOut `json:"prev_out"`
}

type Out struct {
	Type    int    `json:"type"`
	Spent   bool   `json:"spent"`
	Value   int    `json:"value"`
	TxIndex int64  `json:"tx_index"`
	N       int    `json:"n"`
	Script  string `json:"script"`
	Addr    string `json:"addr"`
}

type Transaction struct {
	Hash        string  `json:"hash"`
	Version     int     `json:"ver"`
	VinSize     int     `json:"vin_sz"`
	VoutSize    int     `json:"vout_sz"`
	Size        int     `json:"size"`
	Weight      int     `json:"weight"`
	Fee         int     `json:"fee"`
	RelayedBy   string  `json:"relayed_by"`
	LockTime    int     `json:"lock_time"`
	TxIndex     int64   `json:"tx_index"`
	DoubleSpend bool    `json:"double_spend"`
	Time        int64   `json:"time"`
	BlockIndex  *int64  `json:"block_index"`
	BlockHeight *int64  `json:"block_height"`
	Inputs      []Input `json:"inputs"`
	Out         []Out   `json:"out"`
	Rbf         bool    `json:"rbf"`
}

type PreviousTransaction struct {
	Tx      *Transaction `json:"transaction"`
	TxIndex int64        `json:"tx_index"`
}
type NextTransaction struct {
	Tx      *Transaction `json:"transaction"`
	TxIndex int64        `json:"tx_index"`
}

func GetTransactionByIndex(txIndex int64) (*Transaction, error) {
	// 이전 트랜잭션 조회 URL
	url := fmt.Sprintf("https://blockchain.info/rawtx/%d", txIndex)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching previous transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("transaction not found. Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	}

	// 트랜잭션 데이터 디코딩
	var tx Transaction
	if err := json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return nil, fmt.Errorf("error decoding previous transaction JSON: %v", err)
	}

	return &tx, nil
}

// GetTransactionByAddress 함수를 추가하여 특정 주소의 트랜잭션을 가져옵니다.
func GetTransactionsByAddress(address string) ([]Transaction, error) {
	// 주소 기반 트랜잭션 조회 URL
	url := fmt.Sprintf("https://blockchain.info/address/%s?format=json", address)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching transactions by address: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("transactions not found. Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	}

	// 주소의 트랜잭션 데이터 디코딩
	var addressData struct {
		Transactions []Transaction `json:"txs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&addressData); err != nil {
		return nil, fmt.Errorf("error decoding address transactions JSON: %v", err)
	}

	return addressData.Transactions, nil
}

// CheckTransaction 함수
func CheckTransaction(c echo.Context) error {
	// 클라이언트에서 전달된 트랜잭션 ID 추출
	txid := c.Param("txid")

	// 트랜잭션 정보 요청
	url := fmt.Sprintf("https://blockchain.info/rawtx/%s", txid)
	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error fetching transaction: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return c.String(http.StatusNotFound, fmt.Sprintf("Transaction not found. Status: %d, Response: %s", resp.StatusCode, string(bodyBytes)))
	}

	// 트랜잭션 데이터를 디코딩
	var tx Transaction
	if err := json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error decoding JSON: %v", err))
	}

	// 이전 트랜잭션 정보 추적
	//var previousTxs []PreviousTransaction
	//for _, input := range tx.Inputs {
	//	previousTx, err := GetTransactionByIndex(input.PrevOut.TxIndex)
	//	if err != nil {
	//		fmt.Println("Error fetching previous transaction:", err)
	//		continue
	//	}
	//
	//	previousTxs = append(previousTxs, PreviousTransaction{
	//		Tx:      previousTx,
	//		TxIndex: input.PrevOut.TxIndex,
	//	})
	//}
	// 다음 트랜잭션 정보 추적
	var nextTxs []NextTransaction
	for _, output := range tx.Out {
		nextTransactions, err := GetTransactionsByAddress(output.Addr)
		if err != nil {
			fmt.Println("Error fetching next transactions:", err)
			continue
		}

		for _, nextTx := range nextTransactions {
			nextTxs = append(nextTxs, NextTransaction{
				Tx:      &nextTx,
				TxIndex: nextTx.TxIndex,
			})
		}
	}
	// 응답할 데이터 생성
	responseData := struct {
		Transaction Transaction `json:"transaction"`
		//PreviousTransactions []PreviousTransaction `json:"previous_transactions"`
		NextTransactions []NextTransaction `json:"next_transactions"`
	}{
		Transaction: tx,
		//PreviousTransactions: previousTxs,
		NextTransactions: nextTxs,
	}

	// 현재 트랜잭션 및 이전 트랜잭션 정보 응답
	return c.JSON(http.StatusOK, responseData)
}
