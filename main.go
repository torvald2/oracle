package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// 1. Подключение к базе данных
	dsn := os.Getenv("DB_CONNECTION_STRING")
	if dsn == "" {
		dsn = "postgres://kipinfo_user:kipinfo_password@localhost:5432/kipinfo_db?sslmode=disable"
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Настройка и запуск сервера fasthttp
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			ctx.WriteString("Welcome to Wells Oracle!")
		default:
			if strings.HasPrefix(string(ctx.Path()), "/valuation/") {
				getValuationHandler(ctx)
			} else {
				ctx.Error("Not Found", fasthttp.StatusNotFound)
			}
		}
	}

	log.Println("Starting server on :8080")
	if err := fasthttp.ListenAndServe(":8080", requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func getValuationHandler(ctx *fasthttp.RequestCtx) {
	// Извлекаем well_id из URL. Например, /valuation/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	pathParts := strings.Split(string(ctx.Path()), "/")
	if len(pathParts) != 3 {
		ctx.Error("Invalid URL format. Use /valuation/{well_id}", fasthttp.StatusBadRequest)
		return
	}
	wellIDStr := pathParts[2]

	wellID, err := uuid.Parse(wellIDStr)
	if err != nil {
		ctx.Error("Invalid Well ID format", fasthttp.StatusBadRequest)
		return
	}

	var valuation Valuation
	// Ищем последнюю запись для данного well_id, сортируя по created_at
	result := db.Where("well_id = ?", wellID).Order("created_at desc").First(&valuation)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			ctx.Error("Valuation not found for the given Well ID", fasthttp.StatusNotFound)
		} else {
			ctx.Error("Database error", fasthttp.StatusInternalServerError)
			log.Printf("Database error: %v", result.Error)
		}
		return
	}

	// Сериализуем результат в JSON
	response, err := json.Marshal(valuation)
	if err != nil {
		ctx.Error("Failed to serialize response", fasthttp.StatusInternalServerError)
		log.Printf("JSON marshal error: %v", err)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(response)
}
