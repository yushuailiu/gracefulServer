package main

import (
	"github.com/gin-gonic/gin"
	"time"
	"net/http"
	"github.com/yushuailiu/gracefulServer"
)

func main() {
	route := gin.Default()
	route.GET("/", func(context *gin.Context) {
		time.Sleep(5*time.Second)
		context.String(http.StatusOK, "hello world!")
	})

	graceful := gracefulServer.NewGraceful()


	// add hook before server reload
	graceful.AddBeforeStopHook(func() {
		println("before stop")
	})

	// add hook before server reload
	graceful.AddAfterReloadHook(func() {
		println("after stop")
	})

	// add hook before server reload
	graceful.AddBeforeReloadHook(func() {
		println("before reload")
	})

	// add hook after server reload
	graceful.AddAfterReloadHook(func() {
		println("after reload")
	})

	// set the timeout of the old server when use `kill -USR2 [pid]` reload the server。
	graceful.SetTimeout(20 * time.Second)


	// start the server
	err := graceful.ListenAndServer(":8081", route)

	if err != nil {
		panic(err)
	}

}

