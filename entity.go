package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tealeg/xlsx"
)

// EntityType represents the type of an entity.
type EntityType struct {
	Name   string `json:"name"`
	Marker string `json:"marker"`
}

// Entity represents an entity.
type Entity struct {
	Provider EntityType `json:"provider"`
	Product  EntityType `json:"product"`
}

var entities map[string]string = make(map[string]string)

var entityKind Entity = Entity{
	Provider: EntityType{Name: "provider", Marker: "F"},
	Product:  EntityType{Name: "product", Marker: "P"},
}

// Constants
const (
	PROVIDER_CODE_COLUMN = 1
	PROVIDER_NAME_COLUMN = 2
	PRODUCT_CODE_COLUMN  = 1
	PRODUCT_NAME_COLUMN  = 2
)

func isProvider(eid string) bool {
	return strings.ToUpper(string(eid[0])) == entityKind.Provider.Marker
}

func isProduct(eid string) bool {
	return strings.ToUpper(string(eid[0])) == entityKind.Product.Marker
}

func GetType(eid string) string {
	if isProvider(eid) {
		return entityKind.Provider.Name
	} else if isProduct(eid) {
		return entityKind.Product.Name
	}
	return ""
}

func GetEntities(c echo.Context) error {
	return c.JSON(http.StatusOK, entities)
}

func GetOneEntity(c echo.Context) error {
	eid := c.Param("eid")
	if entityName, exists := entities[eid]; exists {
		entity := map[string]interface{}{
			"ename": entityName,
			"eid":   eid,
			"etype": GetType(eid),
		}
		return c.JSON(http.StatusOK, entity)
	}
	// log.Println(json.Marshal(entities))
	return c.String(http.StatusNotFound, fmt.Sprintf("The entity '%s' does not exist", eid))
}

func GetEntitiesByType(c echo.Context) error {
	etype := c.Param("etype")
	values := []string{}
	entity := Entity{}
	for k, v := range entities {
		if strings.ToLower(etype) == entity.Provider.Name && isProvider(k) {
			values = append(values, v)
		}
		if strings.ToLower(etype) == entity.Product.Name && isProduct(k) {
			values = append(values, v)
		}
	}

	response := map[string][]string{"entities": values}
	return c.JSON(http.StatusOK, response)
}

func LoadConfiguration(filename string) {
	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Log the configuration filename
	log.Printf("Configuration: filename/%s", filename)

	// Open the Excel file
	xlFile, err := xlsx.OpenFile(filename)
	if err != nil {
		log.Fatalf("Error opening Excel file: %v", err)
	}

	// Load Providers data
	log.Println("Configuration: loading/providers/begin")
	loadSheetData(xlFile.Sheet["Fournisseurs"], PROVIDER_CODE_COLUMN, PROVIDER_NAME_COLUMN)
	log.Println("Configuration: loading/providers/end")

	// Load Products data
	log.Println("Configuration: loading/products/begin")
	loadSheetData(xlFile.Sheet["Produits"], PRODUCT_CODE_COLUMN, PRODUCT_NAME_COLUMN)
	log.Println("Configuration: loading/products/end")
}

func loadSheetData(sheet *xlsx.Sheet, codeColumn, nameColumn int) {
	counter := 0
	for _, row := range sheet.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		providerCodeCell := row.Cells[codeColumn-1]
		providerNameCell := row.Cells[nameColumn-1]

		providerCode := providerCodeCell.String()
		providerName := providerNameCell.String()

		if providerCode != "" && providerName != "" {
			entities[strings.ToUpper(providerCode)] = providerName
			log.Println(strings.ToUpper(providerCode), providerName)
			counter++
		}
	}
	log.Printf("Configuration: entities/count/%d", counter)
}