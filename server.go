package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"filippo.io/age"
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

func generateKey() {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	log.Printf("Public key: %s\n", identity.Recipient().String())
	log.Printf("Private key: %s\n", identity.String())
}

func ScriptKeyIdentity() {
	helloWorld := "Hellow World"
	password := "foo12345"
	r, err := age.NewScryptRecipient(password)
	if err != nil {
		log.Fatal(err)
	}
	r.SetWorkFactor(15)
	buf := &bytes.Buffer{}
	w, err := age.Encrypt(buf, r)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.WriteString(w, helloWorld); err != nil {
		log.Fatal(err)
	}

	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	i, err := age.NewScryptIdentity(password)
	if err != nil {
		log.Fatal(err)
	}

	out, err := age.Decrypt(buf, i)
	if err != nil {
		log.Fatal(err)
	}
	outBytes, err := io.ReadAll(out)
	if err != nil {
		log.Fatal(err)
	}
	if string(outBytes) != helloWorld {
		fmt.Errorf("wrong data: %q, excepted %q", outBytes, helloWorld)
	}
}

// func main() {
// 	// generateKey()
// 	// ScriptKeyIdentity()
// }
