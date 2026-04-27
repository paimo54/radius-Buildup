package main

import (
	"log"

	"radius-buildup/internal/config"
	"radius-buildup/internal/database"
	"radius-buildup/internal/models"
)

func main() {
	cfg, _ := config.Load()
	database.Connect(cfg)

	var mapping models.OidMapping
	database.DB.Where("vendor = ?", "zte").First(&mapping)

	log.Printf("ZTE Mapping:")
	log.Printf("OidName: %s", mapping.OidName)
	log.Printf("OidTx: %s", mapping.OidTx)
	log.Printf("OidRx: %s", mapping.OidRx)
	log.Printf("OidDistance: %s", mapping.OidDistance)
}
