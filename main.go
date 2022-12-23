package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/gin-gonic/gin"
	"github.com/plausible-go-clone/migrations"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
	"github.com/uptrace/go-clickhouse/chmigrate"
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
	db := ch.Connect(ch.WithDSN("clickhouse://localhost:8123/test?sslmode=disable"))
	db.AddQueryHook(chdebug.NewQueryHook(
		chdebug.WithEnabled(false),
		chdebug.FromEnv("CHDEBUG"),
	))
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Runs the server",
				Action: func(c *cli.Context) error {
					EventServer()
					return nil
				},
			},

			NewDBcommand(db, chmigrate.NewMigrator(db, migrations.Migrations)),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func NewDBcommand(db *ch.DB, migrator *chmigrate.Migrator) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "Database migrations",
		Subcommands: []*cli.Command{
			{
				Name:  "create_sql",
				Usage: "Create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					fmt.Println("Create command")
					return nil
				},
			},
			{
				Name:  "init",
				Usage: "Create migration tables",
				Action: func(c *cli.Context) error {
					fmt.Println("Create migration")
					return nil
				},
			},
		},
	}

}

func EventServer() {
	router := gin.Default()
	router.POST("/events", HandleEvents)

	router.Run("localhost:8080")
}
