package routes

import (
	"cos/handlers"
	"github.com/labstack/echo/v4"
)

func SetTransactionRoutes(e *echo.Echo) {
	// 트랜잭션 조회 엔드포인트
	e.GET("/transaction/:txid", handlers.CheckTransaction)
}
