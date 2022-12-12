package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Event struct {
	Name string `json:"name"`
}

func HandleEvents(c *gin.Context) {
	var newEvent Event

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newEvent); err != nil {
		return
	}

	c.IndentedJSON(http.StatusOK, newEvent)

}

func main() {

	router := gin.Default()
	router.POST("/events", HandleEvents)

	router.Run("localhost:8080")
}
