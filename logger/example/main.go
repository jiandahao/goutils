package main

import (
	"context"

	"github.com/jiandahao/goutils/logger"
)

func main() {
	l := logger.NewDefaultLogger("info")
	defer l.Sync()

	ctx := context.Background()

	ctx = logger.AppendMetadata(
		ctx,
		logger.NewMetadata().
			Append("request_id", "1234").Append("caller", "main").
			Append("request_id", "5678"),
	)

	l.Infof(ctx, "test info log")

}
