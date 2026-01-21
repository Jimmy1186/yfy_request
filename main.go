package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"kenmec/yfyRequest/jimmy/api"
	"kenmec/yfyRequest/jimmy/config"
	"kenmec/yfyRequest/jimmy/sqlQuery"

	dbs "kenmec/yfyRequest/jimmy/db"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		ctx := c.Request.Context() //每一個rest都只能用獨立的 不要在rest用context.Background()
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
		prettyJSON, _ := json.MarshalIndent(req, "", "  ")

		//fmt.Println("================ DEBUG LOG ================")
		// fmt.Printf("📦 Data Content:\n%s\n", string(prettyJSON))
		// fmt.Println("===========================================")

		scriptName := dbs.CurrentScriptName()
		conveyorFullName := config.Cfg.CONVEYOR_LOCATION_NAME + "-" + scriptName
		conveyorId, err := dbs.Mdb.ConveyorConfigId(ctx, sql.NullString{
			String: conveyorFullName,
			Valid:  true,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"code":      500,
				"message":   "交管尚未初始化",
				"result":    gin.H{},
				"success":   false,
				"timpstamp": timestamp,
			})
			return
		}

		ramdomId, err := uuid.NewV7()
		hId, err := uuid.NewV7()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":      500,
				"message":   "生成 UUID 失敗",
				"result":    gin.H{},
				"success":   false,
				"timpstamp": timestamp,
			})
			return
		}

		err = dbs.WithTransaction(ctx, func(q *sqlQuery.Queries) error {
			err = q.CreateConatiner(ctx, sqlQuery.CreateConatinerParams{
				ID:                    ramdomId.String(),
				CustomID:              sql.NullString{String: req.OrderNo, Valid: true},
				Status:                "AT_LOCATION",
				Owner:                 "CONVEYOR",
				Updatedat:             time.Now(),
				Metadata:              prettyJSON,
				CustomCargoMetadataID: sql.NullString{String: config.Cfg.DEFAULT_CONTAINER_FORMAT_ID, Valid: true},
				ConveyorConfigid:      sql.NullString{String: conveyorId, Valid: true},
			})

			if err != nil {
				return err
			}

			err = q.CreateConatinerHistory(ctx, sqlQuery.CreateConatinerHistoryParams{
				ID:          hId.String(),
				CargoID:     ramdomId.String(),
				Action:      "CREATED",
				Actor:       sql.NullString{String: "ROBOT", Valid: true},
				Description: sql.NullString{String: "Robot arm notify QAMS a container."},
			})
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(500, gin.H{"message": "資料庫操作失敗: " + err.Error()})
			return
		}

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
		ctx := c.Request.Context()
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
		prettyJSON, _ := json.MarshalIndent(req, "", "  ")
		// fmt.Println("================ DEBUG LOG ================")
		// fmt.Printf("📦 Data Content:\n%s\n", string(prettyJSON))
		// fmt.Println("===========================================")

		scriptName := dbs.CurrentScriptName()
		conveyorFullName := config.Cfg.CONVEYOR_LOCATION_NAME + "-" + scriptName
		conveyorId, err := dbs.Mdb.ConveyorConfigId(ctx, sql.NullString{
			String: conveyorFullName,
			Valid:  true,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"code":      500,
				"message":   "交管尚未初始化",
				"result":    gin.H{},
				"success":   false,
				"timpstamp": timestamp,
			})
			return
		}

		ramdomId, err := uuid.NewV7()
		hId, err := uuid.NewV7()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":      500,
				"message":   "生成 UUID 失敗",
				"result":    gin.H{},
				"success":   false,
				"timpstamp": timestamp,
			})
			return
		}

		err = dbs.WithTransaction(ctx, func(q *sqlQuery.Queries) error {
			err = q.CreateConatiner(ctx, sqlQuery.CreateConatinerParams{
				ID:                    ramdomId.String(),
				CustomID:              sql.NullString{String: req.TrayType, Valid: true},
				Status:                "AT_LOCATION",
				Owner:                 "CONVEYOR",
				Updatedat:             time.Now(),
				Metadata:              prettyJSON,
				CustomCargoMetadataID: sql.NullString{String: config.Cfg.DEFAULT_TRAY_FORMAT_ID, Valid: true},
				ConveyorConfigid:      sql.NullString{String: conveyorId, Valid: true},
			})

			if err != nil {
				return err
			}

			err = q.CreateConatinerHistory(ctx, sqlQuery.CreateConatinerHistoryParams{
				ID:          hId.String(),
				CargoID:     ramdomId.String(),
				Action:      "CREATED",
				Actor:       sql.NullString{String: "ROBOT", Valid: true},
				Description: sql.NullString{String: "Robot arm notify QAMS to add tray."},
			})
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(500, gin.H{"message": "資料庫操作失敗: " + err.Error()})
			return
		}

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
