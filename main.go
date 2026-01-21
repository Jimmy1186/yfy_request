package main

import (
	"kenmec/yfyRequest/jimmy/api"

	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	grpcClient := api.NewGRPCClient("localhost:50051")
	go grpcClient.MaintainConnection()
	go grpcClient.StartHeartbeat(30 * time.Second)
	for !grpcClient.IsConnected() {
		log.Println("⏳ 等待 gRPC 連線...")
		time.Sleep(1 * time.Second)
	}

	r := gin.Default()
	r.Use(cors.Default())
	r.Use(func(c *gin.Context) {
		if !grpcClient.IsConnected() {
			timestamp := time.Now().Unix()
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":      500,
				"message":   "gRPC 服務暫時無法使用",
				"result":    gin.H{},
				"success":   false,
				"timpstamp": timestamp,
			})
			c.Abort()
			return
		}
		c.Next()
	})

	armReq := api.NewRobotAmrRequest(grpcClient)

	r.POST("/orderInfo", armReq.OrderInfoReq)

	r.POST("/api/ThirdParty/CallTray", armReq.CallTrayReq)

	r.POST("/api/ThirdParty/Replenish", armReq.Replenish)

	r.Run(":8000")
}
