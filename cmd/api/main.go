package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"

	httpTransport "github.com/freikugel0/boorumesh-be/internal/http"
	"github.com/freikugel0/boorumesh-be/internal/http/handler"
	"github.com/freikugel0/boorumesh-be/internal/repository/postgres"
	"github.com/freikugel0/boorumesh-be/internal/service"
)

func main() {
	// DB
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Repo
	srcRepo := postgres.NewSourceRepositoryPostgres(db)

	// Services
	devSourceSvc := service.NewDevSourceService(srcRepo)
	sourceFetchSvc := service.NewSourceFetchService(srcRepo)

	// Handlers
	devSourceHandler := handler.NewDevSourceHandler(devSourceSvc)
	apiHandler := handler.NewApiHandler(sourceFetchSvc)

	// Router
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := httpTransport.NewRouter(devSourceHandler, apiHandler)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
