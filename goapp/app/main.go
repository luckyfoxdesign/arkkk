package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/iris/v12"
	"github.com/nutsdb/nutsdb"
)

type ValuesData struct {
	Value, Fuid, Tuid int
	FromUnitId        string
	ToUnitId          string
	Status            uint8 // 0 - sww, 1 - correct
	//ErrorSide int // 1 - left (fid), 2 - right (tid), 3 - both
}

type RatiosStruct struct {
	FromUnitRatio float64
	ToUnitRatio   float64
	FormulaID     int32
}
type UnitsIds struct {
	FromUnitId int
	ToUnitId   int
	Status     uint8
}

type UnitStruct struct {
	Ratio     float64
	FormulaID int32
}

func main() {
	bucketName := os.Getenv("BUCKET_NAME")
	mariadburi := os.Getenv("MARDB_URI")

	mardb, err := sql.Open("mysql", mariadburi)
	if err != nil {
		panic(err)
	}

	mardb.SetConnMaxLifetime(time.Minute * 3)
	mardb.SetMaxOpenConns(10)
	mardb.SetMaxIdleConns(10)

	defer mardb.Close()

	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir("./nuts"),
	)
	if err != nil {
		log.Fatal(err)
	}

	app := iris.Default()

	app.Put("/insertkv", func(ctx iris.Context) {
		body, err := ctx.GetBody()
		if err != nil {
			log.Default().Println("body read \n", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}

		if len(body) == 0 {
			log.Default().Println("body empty\n", err)
			ctx.StatusCode(iris.StatusNoContent)
			return
		} else {
			keysWithSpaces := string(body)
			keysSlice := strings.Fields(keysWithSpaces)
			keysToInsert := parseStringValuesToMap(keysSlice)
			notInsertedKeys := insertKeysAndValuesToDB(keysToInsert, db, bucketName)

			if len(notInsertedKeys) > 0 {
				// TODO
				// add not inserted results to the response
				// not only status
				ctx.StatusCode(iris.StatusPartialContent)
				log.Default().Println("not inserted keys:", notInsertedKeys)
				return
			}
			log.Default().Println("all received keys were inserted")
			ctx.StatusCode(iris.StatusOK)
		}
	})

	app.Get("/calc", func(ctx iris.Context) {
		var getStatus uint8 = 1
		var unitsRatios []UnitStruct

		row := ctx.URLParam("r")

		// TODO
		// 1. write logs to db

		parsedValues := parseInputString(row, bucketName, db)
		convertedValues, err := convertUnitsStringIdsToInt(parsedValues)
		if err != nil {
			getStatus = 0
			log.Default().Println("-> convertUnitsStringIdsToInt()\n", err)
		} else {
			unitsRatios, err = returnUnitsRatiosFromDB(mardb, convertedValues.FromUnitId, convertedValues.ToUnitId)
			if err != nil {
				getStatus = 0
				log.Default().Println("-> returnUnitsRatiosFromDB()\n", err)
			}
		}

		ctx.JSON(iris.Map{
			"val":    parsedValues.Value,
			"fur":    unitsRatios[0].Ratio, // from
			"tur":    unitsRatios[1].Ratio, // to
			"fid":    unitsRatios[0].FormulaID,
			"status": getStatus,
		})
	})
	app.Listen(":8080")
}

func insertKeysAndValuesToDB(dataToInsert map[string]string, db *nutsdb.DB, bucketName string) map[string]string {
	notInsertedValues := map[string]string{}
	for k, v := range dataToInsert {
		key := []byte(k)
		value := []byte(v)

		// TODO
		// rewrite the code below with gorutines
		// I can us chan to send not inserted key/value results
		// write them to a struct and convert to json
		if err := db.Update(
			func(tx *nutsdb.Tx) error {
				if err := tx.Put(bucketName, key, value, 0); err != nil {
					return err
				}
				return nil
			}); err != nil {
			notInsertedValues[k] = v
		}
	}
	return notInsertedValues
}

func parseStringValuesToMap(stringToParse []string) map[string]string {
	keyValues := map[string]string{}
	var id string

	for i, v := range stringToParse {
		if _, err := strconv.Atoi(v); err == nil {
			id = v
			continue
		}
		keyValues[stringToParse[i]] = id
		log.Default().Println(v, keyValues[v])
	}

	return keyValues
}

func parseInputString(row, BUCKET_NAME string, db *nutsdb.DB) ValuesData {
	// TODO
	// TOOOO deep nesting for child functions
	var (
		fromValue            int
		fromUnitId, toUnitId string
	)
	valuesData := ValuesData{}
	delimeterIndex := getDelimeter(row)

	if delimeterIndex != -1 {
		partBeforeDelimeter := row[:delimeterIndex]

		fromValue, fromUnitId = parseFromValueAndUnitName(partBeforeDelimeter, BUCKET_NAME, db)

		partAfterDelimeterSlice := strings.Fields(row[delimeterIndex+4:])

		toUnitId = parseToUnitId(partAfterDelimeterSlice, BUCKET_NAME, db)

		valuesData.FromUnitId = fromUnitId
		valuesData.ToUnitId = toUnitId
		valuesData.Value = fromValue
	}

	log.Default().Println("val", fromValue, "fuid:", fromUnitId, "-> tuid:", toUnitId)

	// TODO
	// add an error that says on which side string has an error
	// before delimeter or after
	// just for userfriendly experience with errors on clientside

	if fromUnitId != "" && toUnitId != "" {
		valuesData.Status = 1
	}

	return valuesData
}

func convertUnitsStringIdsToInt(values ValuesData) (UnitsIds, error) {
	var fuid, tuid int
	convertedValues := UnitsIds{}

	fuid, err := strconv.Atoi(values.FromUnitId)
	if err != nil {
		return convertedValues, err
	}
	convertedValues.FromUnitId = fuid

	tuid, err = strconv.Atoi(values.ToUnitId)
	if err != nil {
		return convertedValues, err
	}
	convertedValues.ToUnitId = tuid

	return convertedValues, err
}

func returnUnitsRatiosFromDB(db *sql.DB, fromUnitId, toUnitId int) ([]UnitStruct, error) {
	var result []UnitStruct

	res, err := db.Query("SELECT ratio, formula_id FROM list WHERE unit_id=? || unit_id=?", fromUnitId, toUnitId)
	if err != nil {
		log.Default().Println("-> db.Query()")
		return result, err
	}

	for res.Next() {

		var unit UnitStruct
		err := res.Scan(&unit.Ratio, &unit.FormulaID)

		if err != nil {
			log.Default().Println("-> res.Scan()")
			return result, err
		}

		fmt.Println("UNIT\n", unit)
		result = append(result, unit)
	}
	return result, err
}

func parseToUnitId(partAfterDelimeterSlice []string, BUCKET_NAME string, db *nutsdb.DB) string {
	var toUnitId string
	if len(partAfterDelimeterSlice) > 0 {
		for _, v := range partAfterDelimeterSlice {
			keyValue, err := getUnitIndexFromDB(v, BUCKET_NAME, db)
			if err != nil {
				continue
			}
			if keyValue != "" {
				toUnitId = keyValue
				break
			}
		}
	}
	return toUnitId
}

func getUnitIndexFromDB(unitName, BUCKET_NAME string, db *nutsdb.DB) (string, error) {
	byteName := []byte(unitName)
	var unitIndex string

	err := db.View(
		func(tx *nutsdb.Tx) error {
			e, err := tx.Get(BUCKET_NAME, byteName)
			if err != nil {
				log.Default().Println("tx.Get error:", err)
				return err
			}
			unitIndex = string(e.Value)
			return nil
		})

	if err != nil {
		log.Default().Println("db.View error:", err)
	}
	return unitIndex, err
}

func parseFromValueAndUnitName(fromDataStr, BUCKET_NAME string, db *nutsdb.DB) (int, string) {
	var fromValue, fromValueStartIndex, fromValueEndIndex int = 0, -1, 0
	var fromUnitId string

	for i, v := range fromDataStr {
		if unicode.IsSpace(v) {
			continue
		}
		if unicode.IsDigit(v) && fromValueStartIndex == -1 {
			fromValueStartIndex = i
		}
		if !unicode.IsDigit(v) && fromValueStartIndex != -1 {
			fromValueEndIndex = i
			tmp := strings.ReplaceAll(fromDataStr[fromValueStartIndex:i], " ", "")
			fromValue, _ = strconv.Atoi(tmp)
			break
		}
	}

	valuesArray := strings.Fields(fromDataStr[fromValueEndIndex:])
	fromUnitName := strings.ToLower(valuesArray[0])
	fromUnitId, _ = getUnitIndexFromDB(fromUnitName, BUCKET_NAME, db)

	return fromValue, fromUnitId
}

func getDelimeter(str string) int {
	// TODO
	// store delimeters in different bucket
	var delimetersList = [5]string{" to ", " in ", " в ", " ещ ", " шт "}
	var delimeterIndex int
	for _, v := range delimetersList {
		delimeterIndex = strings.Index(str, v)
		if delimeterIndex != -1 {
			return delimeterIndex
		}
	}
	return -1
}
