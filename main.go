// Package main is the entry point of the application.
package main

import (
	"checklist-api/middleware"
	"checklist-api/routehandlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	r.GET("/checklists", routehandlers.GetChecklists)
	r.GET("/checklist/:id", routehandlers.GetChecklist)
	r.POST("/checklist", routehandlers.PostChecklist)
	r.POST("/checklist/:id/item", routehandlers.PostItem)
	r.PUT("/checklist/:id/item/:itemID", routehandlers.PutItem)

	e := r.Run()

	if e != nil {
		panic(e)
	}
}
