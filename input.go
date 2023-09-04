package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var session *gorm.DB = GetDbSession()

const COMPOST_PLASTIC_CASE_WEIGHT = 1.2 // You can adjust the value accordingly

var RECORDS map[uint]Record = map[uint]Record{}

type Input struct {
	Provider string  `json:"provider"`
	Product  string  `json:"product"`
	Weigth   float64 `json:"weight"`
	Quantity int     `json:"quantity"`
}

func CreateEntryHandler(c echo.Context) error {
	data := Input{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"erreur": err.Error()})
	}

	entry, err := createEntry(c, data)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"erreur": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"id": entry.Id, "poids": entry.Weight})
}

func createEntry(c echo.Context, data Input) (*Record, error) {
	now := time.Now()
	comment := ""

	// Cas spécifique pour le produit "Compost"
	// Réduire le poids de 1,2 kg pour compenser
	// le poids du boîtier en plastique
	weigth := data.Weigth
	if strings.ToLower(data.Product) == "composte" {
		if data.Weigth-COMPOST_PLASTIC_CASE_WEIGHT < 0 {
			return nil, errors.New("Poids trop faible")
		} else {
			weigth = float64(data.Weigth) - COMPOST_PLASTIC_CASE_WEIGHT
			weigth = math.Floor(weigth*100) / 100 // Assurer que nous gardons seulement 2 décimales
		}
	}

	record := Record{
		Provider:  data.Provider,
		Product:   data.Product,
		Weight:    weigth,
		Quantity:  data.Quantity,
		Timestamp: now,
		Comment:   comment,
	}

	if err := session.Create(&record).Error; err != nil {
		return nil, err
	}

	RECORDS[record.Id] = record

	return &record, nil
}

func GetEntries(c echo.Context) error {

	records, _ := GetRecords(session, "")
	return c.JSON(http.StatusOK, records)
}

func DeleteEntry(c echo.Context) error {
	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		return c.JSON(http.StatusNotFound, "Impossible de convertir en rid")
	}
	errDel := DeleteRecord(session, uint(rid))

	if errDel != nil {
		c.JSON(http.StatusOK, fmt.Sprintf("Entree ", rid, " inconnue"))
	}

	return c.JSON(http.StatusOK, "")
}
