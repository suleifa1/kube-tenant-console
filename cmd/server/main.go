package main

import (
	"log"
	"net/http"
	"os"

	"kube-tenant-console/internal/domain"
	"kube-tenant-console/internal/httpapi"
	"kube-tenant-console/internal/kube"
	"kube-tenant-console/internal/local"
)

func main() {
	addr := getenv("ADDR", ":8080")
	dataPath := getenv("DATA_PATH", "./data/state.json")

	store, err := local.OpenStore(dataPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	cfg := domain.Config{
		AllowClusterScope: os.Getenv("ALLOW_CLUSTER_SCOPE") == "true",
	}

	kubeClient, err := kube.NewKubeClientAuto(os.Getenv("KUBECONFIG"))
	if err != nil {
		log.Printf("kubernetes client unavailable: %v", err)
	} else {
		log.Printf("kubernetes client ready")
	}

	server := httpapi.NewServer(store, cfg, kubeClient)
	log.Printf("kube tenant console listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
