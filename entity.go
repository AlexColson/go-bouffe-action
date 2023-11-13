package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

// EntityType represents the type of an entity.
type EntityType struct {
	Name   string `json:"name"`
	Marker string `json:"marker"`
}

// EntitiesDefinition represents an entity.
type EntitiesDefinition struct {
	Provider EntityType `json:"provider"`
	Product  EntityType `json:"product"`
}

type Entity struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

var entities map[string]Entity = make(map[string]Entity)

var entityTypeDefinition EntitiesDefinition = EntitiesDefinition{
	Provider: EntityType{Name: "provider", Marker: "F"},
	Product:  EntityType{Name: "product", Marker: "P"},
}

// Constants
const (
	CODE_COLUMN     = 0
	NAME_COLUMN     = 1
	CATEGORY_COLUMN = 2
)

func isProvider(eid string) bool {
	return strings.ToUpper(string(eid[0])) == entityTypeDefinition.Provider.Marker
}

func isProduct(eid string) bool {
	return strings.ToUpper(string(eid[0])) == entityTypeDefinition.Product.Marker
}

func GetType(eid string) string {
	if isProvider(eid) {
		return entityTypeDefinition.Provider.Name
	} else if isProduct(eid) {
		return entityTypeDefinition.Product.Name
	}
	return ""
}

func GetEntities(c echo.Context) error {
	return c.JSON(http.StatusOK, entities)
}

func GetOneEntity(c echo.Context) error {
	eid := c.Param("eid")
	if entityName, exists := entities[eid]; exists {
		return c.JSON(http.StatusOK, entityName)
	}
	log.Println(json.Marshal(entities))
	return c.String(http.StatusNotFound, fmt.Sprintf("The entity '%s' does not exist", eid))
}

func GetEntitiesByType(c echo.Context) error {
	etype := c.Param("etype")
	// log.Printf("Retrieve entities of type " + etype)
	values := []Entity{}

	for k, v := range entities {
		if strings.ToUpper(etype) == entityTypeDefinition.Provider.Marker && isProvider(k) {
			values = append(values, v)
		}
		if strings.ToUpper(etype) == entityTypeDefinition.Product.Marker && isProduct(k) {
			values = append(values, v)
		}
	}
	vals := values[:]
	sort.Slice(vals, func(i, j int) bool {
		return values[i].Code < values[j].Code
	})
	// log.Println(vals)
	return c.JSON(http.StatusOK, vals)
}

func LoadConfiguration(filename string) {
	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Log the configuration filename
	log.Printf("Configuration: filename/%s", filename)

	// Open the Excel file
	xlFile, err := excelize.OpenFile(filename)
	if err != nil {
		log.Fatalf("Error opening Excel file: %v", err)
	}

	// Load Providers data
	log.Println("Configuration: loading/providers/begin")
	loadSheetData(xlFile, "Fournisseurs")
	log.Println("Configuration: loading/providers/end")

	// Load Products data
	log.Println("Configuration: loading/products/begin")
	loadSheetData(xlFile, "Produits")
	log.Println("Configuration: loading/products/end")
}

func loadSheetData(file *excelize.File, sheetName string) {
	// Get the sheet by name
	sheetIndex, err := file.GetSheetIndex(sheetName)
	if err != nil {
		log.Printf("Sheet not found: %s", sheetName)
		return
	}

	sheet := file.GetSheetName(sheetIndex)

	counter := 0
	rows, _ := file.GetRows(sheet)
	for y, row := range rows {
		if y == 0 {
			continue // Skip the header row
		}
		if len(row) < 2 {
			continue
		}
		entityCode := strings.ToUpper(row[CODE_COLUMN])
		entityName := row[NAME_COLUMN]
		entityCategory := row[CATEGORY_COLUMN]

		// enforce unique codes

		if entityCode != "" && entityName != "" {
			if entity, exists := entities[entityCode]; exists {
				log.Panic("Le code " + entityCode + " existe deja:" + entity.Name)
			}
			entities[entityCode] = Entity{Code: entityCode, Name: entityName, Category: entityCategory}
			counter++
		}
	}
	log.Printf("Configuration: entities/count/%d", counter)
}
