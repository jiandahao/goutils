package main

import (
	"fmt"
	"net/http"

	"github.com/jiandahao/goutils/convhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func testFormData() {
	dasid := "516455"
	dataFname := "The.Falcon.and.The.Winter.Soldier.S01E03.Power.Broker.720p.DSNP.WEB-DL.DDP5.1.H.264-TOMMY/The.Falcon.and.The.Winter.Soldier.S01E03.Power.Broker.720p.DSNP.WEB-DL.DDP5.1.H.264-TOMMY.简体&英文.srt"

	type FileInfo struct {
		Success  bool   `json:"success"`
		FileData string `json:"filedata"`
	}
	resp := FileInfo{}

	formdata := convhttp.NewFormData()
	formdata.Add("dasid", dasid)
	formdata.Add("dafname", dataFname)
	opts := &convhttp.RequestOptions{
		URL:     "https://subhd.tv/ajax/file_ajax",
		Method:  http.MethodPost,
		Request: formdata,
	}

	client := convhttp.DefaultClient
	client.Logger = newLogger()

	if err := client.Do(opts).ShouldBindJSON(&resp); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp.Success)
}

func newLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.Level.UnmarshalText([]byte("debug"))
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	logger, _ := cfg.Build()
	return logger
}

func main() {
	testFormData()
}
