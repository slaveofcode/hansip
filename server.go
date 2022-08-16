package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/securi/repository/pg"
	appRoutes "github.com/slaveofcode/securi/routes"
)

func prepareTmpDirs(dirList []string) error {
	for _, path := range dirList {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(path, os.ModePerm)
			return err
		}
	}
	return nil
}

func main() {
	pgDB := pg.NewRepository(&pg.ConnectionOption{
		DBName: os.Getenv("DATABASE_NAME"),
		Host:   os.Getenv("DATABASE_HOST"),
		Port:   os.Getenv("DATABASE_PORT"),
		User:   os.Getenv("DATABASE_USER"),
		Pass:   os.Getenv("DATABASE_PASSWORD"),
	}, time.UTC)

	if err := pgDB.Connect(context.Background()); err != nil {
		panic(err.Error())
	}

	pgDB.(*pg.RepositoryPostgres).Migrate()
	defer pgDB.Close()

	if err := prepareTmpDirs([]string{
		os.Getenv("UPLOAD_DIR_PATH"),
		os.Getenv("BUNDLED_DIR_PATH"),
	}); err != nil {
		panic("Unable to prepare temp. directories:" + err.Error())
	}

	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	routes := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	routes.Use(cors.New(corsConfig))

	routes.MaxMultipartMemory = 10 << 20 // 10 MiB
	appRoutes.Routes(routes, pgDB.(*pg.RepositoryPostgres))

	server := &http.Server{
		Addr:    os.Getenv("HOSTNAME") + ":" + os.Getenv("PORT"),
		Handler: routes,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
