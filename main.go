package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/swaggo/echo-swagger/example/docs" // docs is generated by Swag CLI, you have to import it.
)

var dataChannel chan ScaleReading = make(chan ScaleReading, 1)

// ServerHeader middleware adds a `Server` header to the response.
func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		return next(c)
	}
}

func NewAppServer() *echo.Echo {
	e, unsecured := InitAppServer("v1")

	// basic ping handler
	unsecured.GET("ping", PingHandler)

	unsecured.GET("scale", Scale)
	unsecured.GET("entities", GetEntities)
	unsecured.GET("entitiesByType/:etype", GetEntitiesByType)
	unsecured.GET("entity/:eid", GetOneEntity)

	unsecured.POST("input", CreateEntryHandler)
	unsecured.GET("input", GetEntries)
	unsecured.GET("input/:date", GetEntries)
	unsecured.DELETE("input/:rid", DeleteEntry)
	unsecured.PUT("input/:rid", UpdateEntry)

	unsecured.GET("download", CreateXLSXFile)
	return e
}

func PingHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, time.Now())
}

func Scale(c echo.Context) error {
	data := ReadScale(dataChannel)
	return c.JSON(http.StatusOK, data)
}

type Config struct {
	UseFakeScale     bool
	UsdScaleDeviceId string
}

func InitAppServer(version string) (*echo.Echo, *echo.Group) {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(ServerHeader)
	e.HideBanner = true

	// e.Use(middleware.Logger())

	notSecured := e.Group(fmt.Sprintf("api/%s/", version))

	e.GET("swagger/*", echoSwagger.WrapHandler)

	static := e.Group("static")
	static.Use(middleware.Static(filepath.Join("static")))
	e.File("/", "index.html")
	e.File("/print", "print.html")

	// e.Logger.SetLevel(loglv1.WARN)
	// e.Use(acecho.LogException())
	return e, notSecured
}

func main() {
	session := GetDbSession()

	// Load produces and providers
	LoadConfiguration("configuration.xlsx")

	// load config if there's one
	var conf Config
	if _, err := toml.DecodeFile("conf.toml", &conf); err != nil {
		log.Println("ERREUR!:Impossible de lire le fichier de configuration conf.toml")
		fmt.Scanln()
		log.Panic()
	}

	if conf.UsdScaleDeviceId == "" {
		log.Println("ERREUR!: Le parametre UsdScaleDeviceId dans conf.toml doit etre defini")
		fmt.Scanln()
		log.Panic()
	}

	// Start the goroutine to read data from the serial port
	if conf.UseFakeScale {
		log.Println("Utilisation d'une balance fictive")
		go FakeScale(dataChannel)
	} else {
		serialPort, err := InitSerial(9600, conf.UsdScaleDeviceId) // Adjust these values
		if err == nil {
			log.Println("ERREUR!: Impossible de communiquer avec la balance")
			fmt.Scanln()
			log.Panic()
		}
		defer serialPort.Close()
		go RealScale(serialPort, dataChannel)

	}

	if session == nil {
		panic("ERREUR!: Impossible de se connecter a la base de donnee")
	}

	e := NewAppServer()

	e.Logger.Fatal(e.Start(":5000"))
}
