package db

import (
	"context"
	"database/sql"
	"kenmec/yfyRequest/jimmy/sqlQuery"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

var DBConn *sql.DB
var Rdb *redis.Client
var Mdb *sqlQuery.Queries

func init() {
	var err error
	dsn := "root:kenmec123@tcp(127.0.0.1:3306)/yfy?parseTime=true"
	DBConn, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("資料庫連線失敗:", err)
	}

	if err := DBConn.Ping(); err != nil {
		log.Fatal("MySQL not reachable:", err)
	}

	Mdb = sqlQuery.New(DBConn)

	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 伺服器位址
		Password: "",               // 如果沒有設定密碼就留空
		DB:       0,                // 使用預設的 DB 0
	})
}

func WithTransaction(ctx context.Context, fn func(q *sqlQuery.Queries) error) error {
	tx, err := DBConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	qtx := Mdb.WithTx(tx)

	err = fn(qtx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
