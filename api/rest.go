package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"kenmec/yfyRequest/jimmy/config"
	dbs "kenmec/yfyRequest/jimmy/db"
	gen "kenmec/yfyRequest/jimmy/protoGen"
	"kenmec/yfyRequest/jimmy/sqlQuery"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RobotAmrRequest struct {
	grpcClient *GRPCClient
}

func NewRobotAmrRequest(grpcClient *GRPCClient) *RobotAmrRequest {
	return &RobotAmrRequest{
		grpcClient: grpcClient,
	}
}

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

func (a *RobotAmrRequest) OrderInfoReq(c *gin.Context) {
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

	curOrder, err := dbs.Mdb.CurrentContainerMaxOrder(ctx, conveyorId)

	if err != nil {
		curOrder = 0
		// c.JSON(http.StatusInternalServerError, gin.H{
		// 	"code":      500,
		// 	"message":   err.Error(),
		// 	"result":    gin.H{},
		// 	"success":   false,
		// 	"timpstamp": timestamp,
		// })
		// return
	}

	err = dbs.WithTransaction(ctx, func(q *sqlQuery.Queries) error {
		err = q.CreateConatiner(ctx, sqlQuery.CreateConatinerParams{
			ID:                    ramdomId.String(),
			CustomID:              sql.NullString{String: req.OrderNo, Valid: true},
			Status:                "AT_LOCATION",
			Owner:                 "CONVEYOR",
			PlacementOrder:        curOrder + 1,
			Updatedat:             time.Now(),
			Metadata:              prettyJSON,
			CustomCargoMetadataID: sql.NullString{String: config.Cfg.DEFAULT_CONTAINER_FORMAT_ID, Valid: true},
			ConveyorConfigID:      sql.NullString{String: conveyorId, Valid: true},
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

	a.grpcClient.SendMessage(&gen.ClientMessage{

		Payload: &gen.ClientMessage_OrderInfo{
			OrderInfo: &gen.OrderInfo{
				OrderNo:     req.OrderNo,
				PaperCount:  *req.PaperCount,
				Height:      *req.Height,
				BatchNo:     req.BatchNo,
				WarehouseNo: req.WarehouseNo,
				TrayType:    req.TrayType,
			},
		},
	})

	// 回傳 JSON 響應
	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"message":   "成功",
		"result":    gin.H{},
		"success":   true,
		"timpstamp": timestamp,
	})

}

func (a *RobotAmrRequest) CallTrayReq(c *gin.Context) {
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
			ConveyorConfigID:      sql.NullString{String: conveyorId, Valid: true},
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

	a.grpcClient.SendMessage(&gen.ClientMessage{
		Payload: &gen.ClientMessage_CallTray{
			CallTray: &gen.CallTray{
				TrayType:    req.TrayType,
				TrayQty:     *req.TrayQty,
				WorkStation: req.WorkStation,
			},
		},
	})

	// 回傳 JSON 響應
	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"message":   "成功",
		"result":    gin.H{},
		"success":   true,
		"timpstamp": timestamp,
	})
}

func (a *RobotAmrRequest) Replenish(c *gin.Context) {
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
	prettyJSON, _ := json.MarshalIndent(req, "", "  ")

	fmt.Println("================ DEBUG LOG ================")
	fmt.Printf("📦 Data Content:\n%s\n", string(prettyJSON))
	fmt.Println("===========================================")

	a.grpcClient.SendMessage(&gen.ClientMessage{
		Payload: &gen.ClientMessage_Replenish{
			Replenish: &gen.Replenish{
				TrayQty:     *req.TrayQty,
				WorkStation: req.WorkStation,
			},
		},
	})

	// 回傳 JSON 響應
	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"message":   "成功",
		"result":    gin.H{},
		"success":   true,
		"timpstamp": timestamp,
	})
}
