package main

import (
	"cos/routes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type Block struct {
	Hash   string `json:"hash"`
	Height int    `json:"height"`
}

func checkNewBlock(c echo.Context) error {
	// 응답을 채우기 위한 channel 생성
	newBlockChan := make(chan string)

	// 고루틴을 통해 블록 확인 시작
	go func() {
		for {
			resp, err := http.Get("https://blockchain.info/latestblock")
			if err != nil {
				fmt.Println("Error:", err)
				time.Sleep(10 * time.Second)
				continue
			}
			defer resp.Body.Close()

			var block Block
			if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
				fmt.Println("Error decoding JSON:", err)
				time.Sleep(10 * time.Second)
				continue
			}

			// 블록 해시와 높이를 채널로 전송
			newBlockChan <- fmt.Sprintf("New Block: %s Height: %d", block.Hash, block.Height)
			time.Sleep(10 * time.Second) // 10초마다 확인
		}
	}()

	// 채널에서 새로운 블록 정보를 수신
	rtnVal := <-newBlockChan
	return c.String(http.StatusOK, rtnVal)
}
func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/check-block", checkNewBlock) // URL 경로 수정
	routes.SetTransactionRoutes(e)

	e.Logger.Fatal(e.Start(":4000"))

}
