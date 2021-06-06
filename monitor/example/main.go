package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiandahao/goutils/monitor"
	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"

	// mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var m *monitor.Monitor

func main() {
	m = monitor.New(&monitor.Config{
		Namespace:      "test_namespace",
		MetricEndpoint: "/metrics",
	})

	monitorHTTPClient()
	monitorHTTPServer()
	monitorDBConntion()

	// register prometheus collector with specified name
	m.RegisterWithName("test_collector",
		prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "test_namespace",
				Name:      "total",
				Help:      "total",
			},
			[]string{"type", "success", "project"},
		),
	)

	// get collector by name
	counter := m.NamedCollector("test_collector").(*prometheus.CounterVec)
	counter.WithLabelValues("message_type", "true", "test_project").Inc()
}

func monitorHTTPClient() {
	httpClient := &http.Client{}
	// set http client monitoring handler,
	// m.SetHTTPClientHandler(/*impl your own http client monitoring handler or use the default one*/nil)

	// wrap and serve http client
	m.WrapAndServeHTTPClient(httpClient)

	httpClient.Get("https://www.baidu.com")
}

func monitorHTTPServer() {
	router := gin.New()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// set http server monitoring handler
	// m.SetHTTPServerHandler(/*impl your own http server monitoring handler or use the default one*/nil)

	// wrap and serve http server
	m.WrapAndServeHTTPServer(httpServer)

	go func() {
		_ = httpServer.ListenAndServe()
	}()
}

func monitorDBConntion() {
	dbConn, _ := gorm.Open("mysql", "root:123@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai")

	// set db connection monitoring handler
	// m.SetDBHandler(/*impl yoyr own db connection monitoring handler or use the default one*/nil)
	// wrap and serve db connection
	m.WrapAndServeDB(dbConn)

	_ = dbConn.Exec("show tables")
}
