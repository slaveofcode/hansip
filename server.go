package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/pg"
	"github.com/slaveofcode/hansip/repository/sqlite"
	appRoutes "github.com/slaveofcode/hansip/routes"
	appConfig "github.com/slaveofcode/hansip/utils/config"
	"github.com/spf13/viper"
)

var awsConfig aws.Config

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

	if appConfig.IsUsingS3Storage() {
		if viper.GetString("aws.region") == "" {
			panic("please set AWS region [config.yaml]")
		}

		if viper.GetString("aws.s3.access_key") == "" || viper.GetString("aws.s3.secret_key") == "" {
			panic("you choose to use S3 storage, please set S3 credentials at the config file [config.yaml]")
		}

		if viper.GetString("aws.s3.bucket_name") == "" {
			panic("please set S3 bucket name [config.yaml]")
		}

		awsCfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(viper.GetString("aws.region")),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					viper.GetString("aws.s3.access_key"),
					viper.GetString("aws.s3.secret_key"),
					"",
				),
			))

		if err != nil {
			panic(fmt.Errorf("unable to load AWS config: %w", err))
		}

		awsConfig = awsCfg
	}

	// Set default config keys
	viper.SetDefault("sqlite.path", "./hansip.db")
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

func getRepo() repository.Repository {
	if viper.GetString("db.type") == "postgresql" {
		return pg.NewRepository(&pg.ConnectionOption{
			DBName: viper.GetString("postgresql.name"),
			Host:   viper.GetString("postgresql.host"),
			Port:   viper.GetString("postgresql.port"),
			User:   viper.GetString("postgresql.user"),
			Pass:   viper.GetString("postgresql.password"),
		}, time.UTC)
	}

	return sqlite.NewRepository(&sqlite.ConnectionOption{
		Path: viper.GetString("sqlite.path"),
	}, time.UTC)
}

func main() {
	readConfig()

	var s3Client *s3.Client
	if appConfig.IsUsingS3Storage() {
		s3Client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {})
	}

	repo := getRepo()

	if err := repo.Connect(context.Background()); err != nil {
		panic(err.Error())
	}

	repo.Migrate()
	defer repo.Close()

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

	appRoutes.Routes(routes, repo, s3Client)

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
