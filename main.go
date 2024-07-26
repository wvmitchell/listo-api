// Package main is the entry point of the application.
package main

import (
	"checklist-api/db/migrate"
	"checklist-api/middleware"
	"checklist-api/routehandlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Load the environment variables
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func main() {
	//Run the migrations
	err := migrate.RunMigrations()
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	// health check
	r.GET(("/"), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Checklist API",
		})
	})

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.AuthMiddleware())

	r.GET("/checklists", routehandlers.GetChecklists)
	r.GET("/checklist/:id", routehandlers.GetChecklist)
	r.PUT("/checklist/:id", routehandlers.PutChecklist)
	r.POST("/checklist", routehandlers.PostChecklist)
	r.DELETE("/checklist/:id", routehandlers.DeleteChecklist)
	r.POST("/checklist/:id/item", routehandlers.PostItem)
	r.PUT("/checklist/:id/items", routehandlers.PutAllItems)
	r.PUT("/checklist/:id/item/:itemID", routehandlers.PutItem)
	r.DELETE("/checklist/:id/item/:itemID", routehandlers.DeleteItem)

	err = r.Run(":80")

	if err != nil {
		panic(err)
	}
}
