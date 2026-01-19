package main

import (
	"encoding/json"
	"fmt"
	"kenmec/yfyRequest/jimmy/api"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type OrderInfo struct {
	//如果是require就必須大寫
	OrderNo     string `json:"orderNo" binding:"required"`
	PaperCount  *int32 `json:"paperCount" binding:"required"`
	Height      *int32 `json:"height" binding:"required"`
	BatchNo     string `json:"batchNo" binding:"required"`
	WarehouseNo string `json:"warehouseNo" binding:"required"`
	TrayType    string `json:"trayType" binding:"required"`
}

type CallTray struct {
	TrayType    string `json:"trayType" binding:"required"`
	TrayQty     *int32 `json:"trayQty" binding:"required"`
	WorkStation string `json:"workStation" binding:"required"`
}

type Replenish struct {
	TrayQty     *int32 `json:"trayQty" binding:"required"`
	WorkStation string `json:"workStation" binding:"required"`
}

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

	r.POST("/orderInfo", func(c *gin.Context) {
		var req OrderInfo
		timestamp := time.Now().Unix()

		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":      500,
				"message":   "請求格式錯誤:" + err.Error(),
				"result":    gin.H{},
				"success":   false,
				"timpstamp": timestamp,
			})
			return
		}

		fmt.Println("================ DEBUG LOG ================")
		prettyJSON, _ := json.MarshalIndent(req, "", "  ")
		fmt.Printf("📦 Data Content:\n%s\n", string(prettyJSON))
		fmt.Println("===========================================")
		// 回傳 JSON 響應
		c.JSON(http.StatusOK, gin.H{
			"code":      200,
			"message":   "成功",
			"result":    gin.H{},
			"success":   true,
			"timpstamp": timestamp,
		})
	})

	r.POST("/api/ThirdParty/CallTray", func(c *gin.Context) {
		var req CallTray
		timestamp := time.Now().Unix()

		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":      500,
				"message":   "請求格式錯誤:" + err.Error(),
				"result":    gin.H{},
				"success":   false,
				"timpstamp": timestamp,
			})
			return
		}

		fmt.Println("================ DEBUG LOG ================")
		prettyJSON, _ := json.MarshalIndent(req, "", "  ")
		fmt.Printf("📦 Data Content:\n%s\n", string(prettyJSON))
		fmt.Println("===========================================")

		// 回傳 JSON 響應
		c.JSON(http.StatusOK, gin.H{
			"code":      200,
			"message":   "成功",
			"result":    gin.H{},
			"success":   true,
			"timpstamp": timestamp,
		})
	})

	r.POST("/api/ThirdParty/Replenish", func(c *gin.Context) {
		var req Replenish
		timestamp := time.Now().Unix()

		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":      500,
				"message":   "請求格式錯誤:" + err.Error(),
				"result":    gin.H{},
				"success":   false,
				"timpstamp": timestamp,
			})
			return
		}

		fmt.Println("================ DEBUG LOG ================")
		prettyJSON, _ := json.MarshalIndent(req, "", "  ")
		fmt.Printf("📦 Data Content:\n%s\n", string(prettyJSON))
		fmt.Println("===========================================")

		// 回傳 JSON 響應
		c.JSON(http.StatusOK, gin.H{
			"code":      200,
			"message":   "成功",
			"result":    gin.H{},
			"success":   true,
			"timpstamp": timestamp,
		})
	})

	r.Run(":8000")
}
