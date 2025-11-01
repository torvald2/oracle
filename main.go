package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/torvald2/wells_oracle/oracle"
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
	srv, err := oracle.NewService("https://sepolia.infura.io/v3/d0c6b5530d5b40c7b43cf44d3ef9bbab", "01b66bfe47a3db46aa7265267e4d382cd261a6e9fe441cc49cb636ce1b1be691")
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Настройка и запуск сервера fasthttp
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch {
		case path == "/":
			ctx.WriteString("Welcome to Wells Oracle!")
		case strings.HasPrefix(path, "/valuation/"):
			getValuationHandler(ctx)
		case strings.HasPrefix(path, "/metadata/"):
			getOpenSeaMetadataHandler(ctx)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
	go func() {
		srv.StartPolling(context.TODO())
	}()
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

func getOpenSeaMetadataHandler(ctx *fasthttp.RequestCtx) {
	// Извлекаем well_id из URL. Например, /opensea/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	pathParts := strings.Split(string(ctx.Path()), "/")
	if len(pathParts) != 3 {
		ctx.Error("Invalid URL format. Use /opensea/{well_id}", fasthttp.StatusBadRequest)
		return
	}
	wellIDStr := pathParts[2]

	wellID, err := uuid.Parse(wellIDStr)
	if err != nil {
		ctx.Error("Invalid Well ID format", fasthttp.StatusBadRequest)
		return
	}

	// 1. Ищем последнюю оценку (valuation) для данного well_id
	var valuation Valuation
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

	// 2. Ищем информацию о скважине (well) по well_id
	var well Well
	result = db.Where("well_id = ?", wellIDStr).First(&well)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			ctx.Error("Well not found for the given Well ID", fasthttp.StatusNotFound)
		} else {
			ctx.Error("Database error", fasthttp.StatusInternalServerError)
			log.Printf("Database error: %v", result.Error)
		}
		return
	}

	// 3. Формируем OpenSea metadata
	metadata := map[string]interface{}{
		"description":  "AI-Powered Dynamic Oil Well Valuation NFT",
		"external_url": "https://openseacreatures.io/3",
		"image":        "ipfs://bafkreid2lid4jiy2hwgvw5abayc6mkvskunb6lfay53pshunjuk5lbfwdm",
		"name":         well.WellName,
		"attributes": []map[string]interface{}{
			{
				"trait_type": "Npv Usd",
				"value":      valuation.NpvUsd,
			},
			{
				"trait_type": "Market Value Usd",
				"value":      valuation.MarketValueUsd,
			},
			{
				"trait_type": "Discount Pct",
				"value":      valuation.DiscountPct,
			},
			{
				"trait_type": "Confidence",
				"value":      valuation.Confidence,
			},
			{
				"trait_type": "Remaining Reserves Bbl",
				"value":      valuation.RemainingReservesBbl,
			},
			{
				"trait_type": "Oil Price Usd",
				"value":      valuation.OilPriceUsd,
			},
			{
				"trait_type": "Operating Cost Per Bbl",
				"value":      valuation.OperatingCostPerBbl,
			},
			{
				"trait_type": "Discount Rate",
				"value":      valuation.DiscountRate,
			},
			{
				"trait_type": "Royalty Rate",
				"value":      valuation.RoyaltyRate,
			},
		},
	}

	// Сериализуем результат в JSON
	response, err := json.Marshal(metadata)
	if err != nil {
		ctx.Error("Failed to serialize response", fasthttp.StatusInternalServerError)
		log.Printf("JSON marshal error: %v", err)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(response)
}
