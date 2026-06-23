package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"admin-aws/internal/agent"
	"admin-aws/internal/config"
	"admin-aws/internal/database"
	httpdelivery "admin-aws/internal/delivery/http"
	"admin-aws/internal/services"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	godotenv.Load()

	config.InitConfig()
	database.InitDB()

	go services.MonitorServers()

	agent.Start()

	router := httpdelivery.SetupRouter()

	domain := os.Getenv("DOMAIN")

	if domain != "" {
		log.Printf("Starting secure server for domain %s (Ports 80/443)", domain)

		if err := os.MkdirAll("certs", 0700); err != nil {
			log.Fatalf("Failed to create certs directory: %v", err)
		}

		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domain),
			Cache:      autocert.DirCache("certs"),
		}

		go func() {
			log.Printf("Starting HTTP-to-HTTPS redirect and ACME listener on port 80")
			log.Fatal(http.ListenAndServe(":80", certManager.HTTPHandler(nil)))
		}()

		server := &http.Server{
			Addr:    ":443",
			Handler: router,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
				MinVersion:     tls.VersionTLS12,
			},
		}

		log.Printf("Starting primary HTTPS server on port 443")
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		port := ":8081"
		log.Printf("DOMAIN not set in .env. Starting standard backend server on port %s", port)
		log.Fatal(http.ListenAndServe(port, router))
	}
}
