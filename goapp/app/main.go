package main

import (
	"database/sql"

	//"fmt"
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

type KeyValuePair struct {
	Key string
	Val string
}

func main() {
	var nutsDbPath string = "./nuts"

	unitsBucketName := os.Getenv("UNITS_BUCKET_NAME")
	mariadburi := os.Getenv("MARDB_URI")

	files, err := os.ReadDir("./nuts_tmp")
	if err != nil {
		log.Default().Println("-> os.ReadDir()", err)
	}

	if len(files) > 0 {
		nutsDbPath = "./nuts_tmp"
	}

	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(nutsDbPath),
	)
	if err != nil {
		log.Fatal(err)
	}

	mardb, err := sql.Open("mysql", mariadburi)
	if err != nil {
		panic(err)
	}

	mardb.SetConnMaxLifetime(time.Minute * 3)
	mardb.SetMaxOpenConns(10)
	mardb.SetMaxIdleConns(10)

	defer mardb.Close()

	app := iris.Default()

	app.Put("/insertkv", func(ctx iris.Context) {
		body, err := ctx.GetBody()
		if err != nil {
			log.Default().Println("-> app.Put(/insertkv) -> body read \n", err)
			ctx.StatusCode(iris.StatusInternalServerError)
			return
		}

		if len(body) == 0 {
			log.Default().Println("-> app.Put(/insertkv) -> body empty\n", err)
			ctx.StatusCode(iris.StatusNoContent)
			return
		} else {
			keysWithSpaces := string(body)
			keysSlice := strings.Fields(keysWithSpaces)
			keysToInsert := parseStringValuesToMap(keysSlice)
			notInsertedKeys := insertKeysAndValuesToDB(keysToInsert, db, unitsBucketName)

			if len(notInsertedKeys) > 0 {
				// TODO
				// add not inserted results to the response
				// not only status
				ctx.StatusCode(iris.StatusPartialContent)
				log.Default().Println("-> app.Put(/insertkv) -> not inserted keys:", notInsertedKeys)
				return
			}
			log.Default().Println("-> app.Put(/insertkv) -> all received keys were inserted")
			ctx.StatusCode(iris.StatusOK)
		}
	})

	app.Get("/keys/list", func(ctx iris.Context) {
		bucketKeyValuesList := getAllBucketKeyPairs(unitsBucketName, db)
		responseJSON := iris.Map{"list": bucketKeyValuesList}
		responseOptions := iris.JSON{Indent: "", Secure: true}
		ctx.JSON(responseJSON, responseOptions)
	})

	app.Get("/calc", func(ctx iris.Context) {
		var getStatus uint8 = 1
		var unitsRatios []UnitStruct

		row := ctx.URLParam("r")

		// TODO
		// 1. write logs to db

		parsedValues := parseInputString(row, unitsBucketName, db)
		convertedValues, err := convertUnitsStringIdsToInt(parsedValues)
		if err != nil {
			getStatus = 0
			log.Default().Println("-> app.Get(/calc) -> convertUnitsStringIdsToInt()\n", err)
		}

		if convertedValues.Status != 0 {
			unitsRatios, err = returnUnitsRatiosFromDB(mardb, convertedValues.FromUnitId, convertedValues.ToUnitId)
			if err != nil {
				getStatus = 0
				log.Default().Println("-> app.Get(/calc) -> returnUnitsRatiosFromDB()\n", err)
			}
		}

		responseJSON := returnResponseIrisMap(unitsRatios)
		responseJSON["status"] = getStatus
		responseJSON["val"] = parsedValues.Value

		responseOptions := iris.JSON{Indent: "", Secure: true}

		ctx.JSON(responseJSON, responseOptions)
	})
	app.Listen(":8080")
}

// Заполняем карту данными для ответа клиенту из полученной структуры с данными из базы
func returnResponseIrisMap(unitsRatios []UnitStruct) iris.Map {
	var responseJSON iris.Map = iris.Map{
		"val":    0,
		"fur":    0,
		"tur":    0,
		"fid":    0,
		"status": 0,
	}

	checkStatus := checkParsedUnitsRatiosAndFormulaExists(unitsRatios)

	if checkStatus {
		responseJSON["fur"] = unitsRatios[0].Ratio
		responseJSON["tur"] = unitsRatios[1].Ratio
		responseJSON["fid"] = unitsRatios[1].FormulaID
		responseJSON["status"] = 1
	}

	return responseJSON
}

// Проверка наличия в структуре обоих значение from/to
// смысла отправлять частями кажется нет
func checkParsedUnitsRatiosAndFormulaExists(units []UnitStruct) bool {
	var flag bool

	if len(units) == 2 {
		flag = true
	}

	return flag
}

// Перебор входной карты и добавление в локальную базу пар ключ-значение
// Можно использовать как для массового добавления, так и для одного
// Используется для админки
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

// Перебор значений массива и добавление этих значений в карту
// Формат даных: 102 one two three 104 five six seven ...
// Используется для админки
func parseStringValuesToMap(stringToParse []string) map[string]string {
	keyValues := map[string]string{}
	var id string

	for i, v := range stringToParse {
		if _, err := strconv.Atoi(v); err == nil {
			id = v
			continue
		}
		keyValues[stringToParse[i]] = id
		//log.Default().Println(v, keyValues[v])
	}

	return keyValues
}

func parseInputString(row, unitsBucketName string, db *nutsdb.DB) ValuesData {
	// TODO
	// TOOOO deep nesting for child functions
	var (
		fromValue            int
		fromUnitId, toUnitId string
	)
	valuesData := ValuesData{}
	delimeterIndex := getDelimeter(row, unitsBucketName, db)

	if delimeterIndex != -1 {
		partBeforeDelimeter := row[:delimeterIndex]

		// для from unit всегда будет значение т.к мы берем все что левее разделителя
		fromValue, fromValueEndIndex := parseFromValue(partBeforeDelimeter)
		fromUnitId := parseFromUnitId(partBeforeDelimeter, unitsBucketName, fromValueEndIndex, db)

		partAfterDelimeterSlice := strings.Fields(row[delimeterIndex+4:])
		toUnitId = parseToUnitId(partAfterDelimeterSlice, unitsBucketName, db)

		// добавить функцию отправки в базу того что в итоге вышло для конвертации

		// значения id могут быть пустыми если они не были найдены в базе
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

// Конвертация строчного значения в целочисленное

// Т.к из базы самый простой вариант вернуть значение в виде строки
// то были мысли что конвертация в число может быть эффективнее на этом этапе
// нежели mysql/maria будет занимается приведением типов на своей стороне

// Могу быть неправ, но знаний у меня недостаточно. Есть ощущения что функция тяжелая, а выполняет ненужную работу.
func convertUnitsStringIdsToInt(values ValuesData) (UnitsIds, error) {
	var fuid, tuid int
	convertedValues := UnitsIds{}

	if values.FromUnitId == "" || values.ToUnitId == "" {
		return convertedValues, nil
	}

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

// Запрос в mysql/maria где на основе id элементов запрашиваются их коэффициенты по отношению к базовой единице
func returnUnitsRatiosFromDB(db *sql.DB, fromUnitId, toUnitId int) ([]UnitStruct, error) {
	var result []UnitStruct

	res, err := db.Query("SELECT ratio, formula_id FROM list WHERE unit_id=? || unit_id=?", fromUnitId, toUnitId)
	if err != nil {
		log.Default().Println("-> returnUnitsRatiosFromDB() -> db.Query()")
		return result, err
	}

	for res.Next() {

		var unit UnitStruct
		err := res.Scan(&unit.Ratio, &unit.FormulaID)

		if err != nil {
			log.Default().Println("-> returnUnitsRatiosFromDB() -> res.Scan()")
			return result, err
		}

		result = append(result, unit)
	}
	return result, err
}

// Функция перебирает значения из входного массива и проверяет каждый ключ на его наличие в базе
// Будет возврашен первый найденный ключ. При его отсутсвии, вернется пустая строка
// Например есть строка 1unit to {nothing do I unit}
// Входной массив из элементов {nothing do I unit} будет перебираться до тех пор пока в базе не найдется соответствующего ключа
// или значения в массиве не закончится. Из массива найден будет {unit}
func parseToUnitId(partAfterDelimeterSlice []string, bucketName string, db *nutsdb.DB) string {
	var toUnitId string
	if len(partAfterDelimeterSlice) > 0 {
		for _, v := range partAfterDelimeterSlice {
			keyValue, err := getUnitIndexFromDB(v, bucketName, db)
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

// Функция по ключу запрашивает значение этого ключа из локальной базы данных
// Если ключа нет, функция вернет пустое значение
func getUnitIndexFromDB(unitName, bucketName string, db *nutsdb.DB) (string, error) {
	byteName := []byte(unitName)
	var unitIndex string

	err := db.View(
		func(tx *nutsdb.Tx) error {
			e, err := tx.Get(bucketName, byteName)
			if err != nil {
				log.Default().Println("-> getUnitIndexFromDB() -> tx.Get error:", err)
				return err
			}
			unitIndex = string(e.Value)
			return nil
		})

	if err != nil {
		log.Default().Println("-> getUnitIndexFromDB() -> db.View error:", err)
	}
	return unitIndex, err
}

// Функция разбирает входную строку на массив и берет первое значение
// Например полная строка 123{unit some another word} to unit
// Из входной строки вида {unit some another word} функция возьмет первое значение {unit}
func parseFromUnitId(sliceBetweenValueAndDelimeter, bucketName string, fromValueEndIndex int, db *nutsdb.DB) string {

	valuesArray := strings.Fields(sliceBetweenValueAndDelimeter[fromValueEndIndex:])
	//TODO
	// could we analize all values from slice to understand which value is correct??
	fromUnitName := strings.ToLower(valuesArray[0])
	fromUnitId, _ := getUnitIndexFromDB(fromUnitName, bucketName, db)

	return fromUnitId
}

// В входной строке ищет числа
// Возращает само число если такое есть и индекс последнего символа этого значения
// Например если строка равна {123unit to unit} то функция вернет {123} - число, {2}  индекс (отсчет от 0)
func parseFromValue(strPartBeforeDelimeter string) (int, int) {
	var fromValue, fromValueStartIndex, fromValueEndIndex int = 0, -1, 0

	for i, v := range strPartBeforeDelimeter {
		if unicode.IsSpace(v) {
			continue
		}
		if unicode.IsDigit(v) && fromValueStartIndex == -1 {
			fromValueStartIndex = i
		}
		if !unicode.IsDigit(v) && fromValueStartIndex != -1 {
			fromValueEndIndex = i
			tmp := strings.ReplaceAll(strPartBeforeDelimeter[fromValueStartIndex:i], " ", "")
			fromValue, _ = strconv.Atoi(tmp)
			break
		}
	}

	return fromValue, fromValueEndIndex
}

// Поиск в входной строке разделителя
// Возвращает индекс первого символа разделителя
func getDelimeter(str, bucketName string, db *nutsdb.DB) int {
	// TODO
	// store delimeters in different bucket
	var delimetersList = [5]string{" to ", " in ", " в ", " ещ ", " шт "}
	var delimeterIndex int = -1
	for _, v := range delimetersList {
		delimeterIndex = strings.Index(str, v)
		if delimeterIndex != -1 {
			return delimeterIndex
		}
	}

	// подумал что я слишком упоролся прикручивать запрос в базу чтобы получить пару значений
	// пока код оставлю, вдруг в будущем понадобится

	/* Метод запроса в базу с получением всех ключей, перебором ключей с поиском по строке
	// err := db.View(
	// 	func(tx *nutsdb.Tx) error {
	// 		entries, err := tx.GetAll(bucketName)
	// 		if err != nil {
	// 			log.Default().Println("-> getDelimeter() -> tx.GetAll error:", err)
	// 			return err
	// 		}

	// 		for _, entry := range entries {
	// 			key := string(entry.Key)
	// 			keyIndex := strings.Index(str, key)

	// 			if keyIndex != -1 {
	// 				delimeterIndex = keyIndex
	// 			}
	// 		}

	// 		return nil
	// 	})

	// if err != nil {
	// 	log.Default().Println("-> getDelimeter() -> db.View error:", err)
	// }
	*/

	return delimeterIndex
}

// Получение всех сохраненных пар ключ-значение из базы
// Используется для отображения списка в админке
func getAllBucketKeyPairs(bucketName string, db *nutsdb.DB) []KeyValuePair {
	var bucketKeyValuesList []KeyValuePair

	err := db.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(bucketName)
			if err != nil {
				log.Default().Println("-> getAllBucketKeyPairs() -> tx.GetAll error:", err)
				return err
			}

			for _, entry := range entries {
				pair := KeyValuePair{}
				pair.Key = string(entry.Key)
				pair.Val = string(entry.Value)
				bucketKeyValuesList = append(bucketKeyValuesList, pair)
			}

			return nil
		})

	if err != nil {
		log.Default().Println("-> getAllBucketKeyPairs() -> db.View error:", err)
	}

	return bucketKeyValuesList
}
