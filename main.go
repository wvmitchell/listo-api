// Package main is the entry point of the application.
package main

import (
	"checklist-api/db/migrate"
	"checklist-api/middleware"
	"checklist-api/routehandlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Run the migrations
	err := migrate.RunMigrations()

	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	r.GET("/checklists", routehandlers.GetChecklists)
	r.GET("/checklist/:id", routehandlers.GetChecklist)
	r.POST("/checklist", routehandlers.PostChecklist)
	r.POST("/checklist/:id/item", routehandlers.PostItem)
	r.PUT("/checklist/:id/item/:itemID", routehandlers.PutItem)

	err = r.Run()

	if err != nil {
		panic(err)
	}
}
