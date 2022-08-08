package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	appRoutes "github.com/slaveofcode/securi/routes"
)

func prepareUploadedDir() error {
	path := os.Getenv("UPLOAD_DIR_PATH")
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		return err
	}

	return nil
}

func main() {
	if err := prepareUploadedDir(); err != nil {
		panic("Unable to create uploaded directory:" + err.Error())
	}

	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	routes := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	routes.Use(cors.New(corsConfig))

	routes.MaxMultipartMemory = 10 << 20 // 10 MiB
	appRoutes.Routes(routes)

	server := &http.Server{
		Addr:    os.Getenv("HOSTNAME") + ":" + os.Getenv("PORT"),
		Handler: routes,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
