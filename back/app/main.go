package main

import (
"fmt"
"strconv"
	"strings"
	"unicode"

	//"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

func main() {

	var unitsCodes = map[string]int{
		"x": 105,
		"y": 105,
	}
	app := iris.Default()

	// Enable CORS
	// corsConfig := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"http://localhost:3000"}, // Adjust the origin as per your frontend app's URL
	// 	AllowedMethods:   []string{"GET", "POST"},
	// 	// AllowedHeaders:   []string{"Origin", "Content-Type", "Accept"},
	// 	AllowCredentials: true,
	// })
	//app.Use(corsConfig)

	app.Get("/calc", func(ctx iris.Context) {
		row := ctx.URLParam("r")

		fromValue, fid, tid := test(row, &unitsCodes)
		ctx.JSON(iris.Map{
			"val": fromValue,
			"fid": fid,
			"tid": tid,
		})
	})
	app.Listen(":8080")
}

func test(row string, unitsCodes *map[string]int) (int, int, int) {
	var (
		fromValue,     tid            int
		fromUnitName, toUnitName string
	)
	delimeterIndex, _ := getDelimeter(row)
		fmt.Println("initial row:",row)

		if delimeterIndex != -1 {
			partBeforeDelimeter := row[:delimeterIndex]
			fromValue, fromUnitName = parseFromDataSlice(partBeforeDelimeter)
			partAfterDelimeterSlice := strings.Fields(row[delimeterIndex+4:])

			if len(partAfterDelimeterSlice) > 0 {
				for _, v := range partAfterDelimeterSlice {
					toUnitName = v
					k := (*unitsCodes)[v]
					if k != 0 {
						tid = k
						break
					}
				}
			}
		}

		fromUnitName = strings.ToLower(fromUnitName)
		toUnitName = strings.ToLower(toUnitName)

		fmt.Println(fromValue, "fun:",fromUnitName, "-> tun:", toUnitName)

		fid := (*unitsCodes)[fromUnitName]

		return fromValue, fid, tid
}

func parseFromDataSlice(fromDataStr string) (int, string) {
	var fromValue, fromValueStartIndex, fromValueEndIndex int = 0, -1, 0
	var fromValueUnitName string

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

	t := strings.Fields(fromDataStr[fromValueEndIndex:])
	fromValueUnitName = t[0]

	return fromValue, fromValueUnitName
}

func getDelimeter(str string) (int, string) {
	var delimetersList = [5]string{" to ", " in ", " в ", " ещ ", " шт "}
	var delimeterIndex int
	for _, v := range delimetersList {
		delimeterIndex = strings.Index(str, v)
		if delimeterIndex != -1 {
			return delimeterIndex, v
		}
	}
	return -1, ""
}
