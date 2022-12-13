package main

import (
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"

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

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Runs the server",
				Action: func(cCtx *cli.Context) error {
					EventServer()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func EventServer() {
	router := gin.Default()
	router.POST("/events", HandleEvents)

	router.Run("localhost:8080")
}
