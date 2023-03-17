package main

import (
	"fmt"
	"net/http"

	"github.com/idproxy/httpserver/pkg/hctx"
	"github.com/idproxy/httpserver/pkg/middleware/logger"
	"github.com/idproxy/httpserver/pkg/server"
)

var db = make(map[string]string)

func main() {
	s := server.New()
	s.Router().Use(logger.Logger())
	r := s.Router()

	r.GET("/", func(hctx hctx.Context) {
		fmt.Println("ping")
		hctx.String(http.StatusOK, "pang")
	})

	r.GET("/ping", func(hctx hctx.Context) {
		hctx.String(http.StatusOK, "pong")
	})

	r.GET("/ping/pong/pang", func(hctx hctx.Context) {
		hctx.String(http.StatusOK, "ping.pong.pang")
	})

	testRouter := r.Group("/test")
	testRouter.GET("/test1", func(hctx hctx.Context) {
		hctx.JSON(http.StatusOK, "test1")
	})
	testRouter.GET("/test2", func(hctx hctx.Context) {
		hctx.JSON(http.StatusOK, "test2")
	})

	r.GET("/user/:name", func(hctx hctx.Context) {
		params := hctx.GetParams()
		fmt.Println(params)
		user, _ := params.Get("name")
		fmt.Println(user)
		value, ok := db[user]
		if ok {
			hctx.JSON(http.StatusOK, map[string]any{"user": user, "value": value})
		} else {
			hctx.JSON(http.StatusOK, map[string]any{"user": user, "status": "no value"})
		}
	})

	r.POST("/user/:name", func(hctx hctx.Context) {
		user, _ := hctx.GetParams().Get("name")
		fmt.Println(user)
		db[user] = "ok"
		hctx.JSON(http.StatusOK, "")
	})

	// Get user value
	r.GET("/user/:name/:provider", func(hctx hctx.Context) {
		provider, _ := hctx.GetParams().Get("provider")
		value, ok := db[provider]
		if ok {
			hctx.JSON(http.StatusOK, map[string]any{"provider": provider, "value": value})
		} else {
			hctx.JSON(http.StatusOK, map[string]any{"provider": provider, "status": "no value"})
		}
	})

	r.GET("/user/:name/:surname/status", func(hctx hctx.Context) {
		name, _ := hctx.GetParams().Get("name")
		surname, _ := hctx.GetParams().Get("surnmae")
		fmt.Println(name)
		fmt.Println(surname)
		hctx.JSON(http.StatusOK, "")
	})

	r.GET("/a/////", func(hctx hctx.Context) {
		hctx.String(http.StatusOK, "pong")
	})

	s.PrintRoutes()
	s.Run(":8889")
}
