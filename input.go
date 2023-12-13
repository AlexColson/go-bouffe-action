package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var session *gorm.DB = GetDbSession()

var COMPOST_PLASTIC_CASE_WEIGHT = 1.2 // You can adjust the value accordingly

var RECORDS map[uint]Record = map[uint]Record{}

type Input struct {
	Provider string  `json:"provider"`
	Product  string  `json:"product"`
	Category string  `json:"category"`
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

	return c.JSON(http.StatusOK, map[string]interface{}{"id": entry.Id, "weight": entry.Weight})
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
			return nil, errors.New("Poids trop faible, doit etre superieur a :" + strconv.FormatFloat(COMPOST_PLASTIC_CASE_WEIGHT, 'f', -1, 64) + "kg")
		} else {
			weigth = float64(data.Weigth) - COMPOST_PLASTIC_CASE_WEIGHT
			weigth = math.Floor(weigth*100) / 100 // Assurer que nous gardons seulement 2 décimales
		}
	}

	record := Record{
		Provider:  data.Provider,
		Product:   data.Product,
		Category:  data.Category,
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
	filterDate := c.Param("date")
	records, _ := GetRecords(session, filterDate)
	return c.JSON(http.StatusOK, records)
}

func DeleteEntry(c echo.Context) error {
	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		return c.JSON(http.StatusNotFound, "Impossible de convertir en rid")
	}
	errDel := DeleteRecord(session, uint(rid))

	if errDel != nil {
		c.JSON(http.StatusOK, fmt.Sprintf("Entree %d inconnue", rid))
	}

	return c.JSON(http.StatusOK, "")
}

func UpdateEntry(c echo.Context) error {
	rid, _ := strconv.Atoi(c.Param("rid"))
	m := echo.Map{}
	err := c.Bind(&m)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	quantity := 0
	comment := ""
	if m["quantity"] != nil {
		quantity = int(m["quantity"].(float64))
	}
	if m["comment"] != nil {
		comment = m["comment"].(string)
	}

	delerr := UpdateRecord(session, uint(rid), quantity, comment)
	if delerr != nil {
		return c.JSON(http.StatusBadRequest, delerr.Error())
	}
	return c.JSON(http.StatusOK, "")
}

func CreateXLSXFile(c echo.Context) error {
	// Create a new Excel file
	f := excelize.NewFile()
	defer f.Close()
	// Create a new sheet
	today := time.Now().Format("2006-01-02")
	sheetName := "Saisies"
	index, err := f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")
	// Set active sheet of the Excel file
	f.SetActiveSheet(index)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create XLSX sheet"})
	}

	// Set the column headers
	f.SetCellValue(sheetName, "A1", "#Date")
	f.SetCellValue(sheetName, "B1", "Fournisseur")
	f.SetCellValue(sheetName, "C1", "Categorie")
	f.SetCellValue(sheetName, "D1", "Produit")
	f.SetCellValue(sheetName, "E1", "Poids")
	f.SetCellValue(sheetName, "F1", "Remarques")

	// Get the date in the desired format

	// records, err := GetRecords(session, "")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get records"})
	}

	// Run the grouped query
	var result []struct {
		Date     string  `json:"date"`
		Provider string  `json:"provider"`
		Category string  `json:"category"`
		Product  string  `json:"product"`
		Comment  string  `json:"comment"`
		Total    float64 `json:"total"`
	}

	session.Table("records").
		Select("strftime('%Y-%m-%d', timestamp) as date, provider, category, product, comment, sum(quantity * weight) as total").
		Group("date, provider, category, product").
		Order("date, provider, category, product").
		Scan(&result)

	log.Println("Export: begin")
	row := 2
	for _, record := range result {

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), record.Date)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), record.Provider)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), record.Category)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), record.Product)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), record.Total)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), record.Comment)
		row++

	}
	log.Println("Export: end")

	// Save the Excel file
	dir := os.TempDir()
	name := fmt.Sprintf("data_%s.xlsx", today)
	filename := filepath.Join(dir, name)

	f.Close()
	log.Println("Saving file to " + filename)
	if err := f.SaveAs(filename); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create XLSX file"})
	}
	defer os.Remove(filename)
	// Send the file back to the browser as an attachment
	return c.Attachment(filename, name)
}
