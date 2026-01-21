package db

import (
	"context"
	"strconv"
)

func CurrentScriptId() string {

	ctx := context.Background()
	txt, err := Rdb.Get(ctx, "current-script-id").Result()

	if err == nil {
		scriptId, _ := strconv.Unquote(txt)
		return scriptId
	}

	dbData, err := Mdb.CurrentScriptId(ctx)

	return dbData
}

func CurrentScriptName() string {
	ctx := context.Background()
	txt, err := Rdb.Get(ctx, "current-script-name").Result()

	if err == nil {
		scriptName, _ := strconv.Unquote(txt)
		return scriptName
	}

	dbData, err := Mdb.CurrentScriptName(ctx)
	return dbData
}
