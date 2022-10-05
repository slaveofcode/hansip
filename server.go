package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository/pg"
	appRoutes "github.com/slaveofcode/hansip/routes"
	"github.com/spf13/viper"
)

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")             // working directory
	viper.AddConfigPath("$HOME/.hansip") // hansip app directory
	viper.AddConfigPath("/etc/hansip")   // system directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("please create the config file [config.yaml]"))
		} else {
			panic(fmt.Errorf("unable to read config file [config.yaml]: %w", err))
		}
	}

	// Set default config keys
	viper.SetDefault("server_api.host", "localhost")
	viper.SetDefault("site.shortlink_path", "/d")
}

func prepareDirs(dirList []string) error {
	for _, path := range dirList {
		var err error
		if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				log.Println("Unable create TMP dir:", err.Error())
				return err
			}

		}
	}
	return nil
}

func main() {
	readConfig()

	pgDB := pg.NewRepository(&pg.ConnectionOption{
		DBName: viper.GetString("database.name"),
		Host:   viper.GetString("database.host"),
		Port:   viper.GetString("database.port"),
		User:   viper.GetString("database.user"),
		Pass:   viper.GetString("database.password"),
	}, time.UTC)

	if err := pgDB.Connect(context.Background()); err != nil {
		panic(err.Error())
	}

	pgDB.(*pg.RepositoryPostgres).Migrate()
	defer pgDB.Close()

	if err := prepareDirs([]string{
		viper.GetString("dirpaths.upload"),
		viper.GetString("dirpaths.bundle"),
	}); err != nil {
		panic("Unable to prepare temp. directories:" + err.Error())
	}

	gin.SetMode(gin.ReleaseMode)
	routes := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Disposition"}
	routes.Use(cors.New(corsConfig))
	routes.Use(gin.Recovery())

	routes.MaxMultipartMemory = viper.GetInt64("upload.max_mb") << 20
	appRoutes.Routes(routes, pgDB.(*pg.RepositoryPostgres))

	serverAddr := viper.GetString("server_api.host") + ":" + viper.GetString("server_api.port")

	server := &http.Server{
		Addr:    serverAddr,
		Handler: routes,
	}

	log.Println("Server Started at:", fmt.Sprintf("http://%s", serverAddr))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
