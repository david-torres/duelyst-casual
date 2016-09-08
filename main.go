package main

import (
	"log"

	r "github.com/dancannon/gorethink"
	"github.com/david-torres/duelyst-casual/controllers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	middleware "github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
)

// Config allows config values can be accessed
var Config = viper.New()

func init() {
	// read in configurations
	Config.SetConfigName("config")
	Config.AddConfigPath("$GOPATH/src/github.com/david-torres/duelyst-casual/configs")
	err := Config.ReadInConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func main() {
	// init the web server
	e := echo.New()

	// init app-wide middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	// security measures
	e.Use(middleware.BodyLimit("2M")) // sets maximum request body size
	e.Use(middleware.CSRF())          // default protection against session riding
	e.Use(middleware.Secure())        // default protection against XSS, content sniffing, clickjacking, and other infections

	// init static assets
	e.Static("/assets", "assets")

	// init db
	host := Config.GetString("db.host")
	database := Config.GetString("db.database")

	session, err := r.Connect(r.ConnectOpts{
		Address:  host,
		Database: database,
	})

	result, err := r.DBList().Contains(database).Run(session)
	if err != nil {
		log.Fatalf("error checking if database exists %s", err)
	}

	var dbExists bool
	result.One(&dbExists)

	if !dbExists {
		log.Println("Creating db and tables")
		r.DBCreate(database).RunWrite(session)
		r.TableCreate("games").RunWrite(session)
	}

	// routes
	e.File("/", "public/index.html")
	e.GET("/ws", standard.WrapHandler(controllers.Socket(session)))

	// start the server
	e.Run(standard.New(":3000"))
}
