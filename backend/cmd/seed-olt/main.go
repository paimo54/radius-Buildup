package main

import (
	"log"
	"radius-buildup/internal/config"
	"radius-buildup/internal/database"
	"radius-buildup/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	_, err = database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Auto migrate tables
	err = database.DB.AutoMigrate(&models.Olt{}, &models.OidMapping{})
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	// Insert mappings
	sql := `INSERT IGNORE INTO oid_mappings (vendor, oid_name, oid_tx, oid_rx, oid_status) VALUES
('hioso', '.1.3.6.1.4.1.25355.3.2.6.3.2.1.37', '.1.3.6.1.4.1.25355.3.2.6.14.2.1.4', '.1.3.6.1.4.1.25355.3.2.6.14.2.1.8', '.1.3.6.1.4.1.25355.3.2.6.3.2.1.39'),
('hsgq', '.1.3.6.1.4.1.50224.3.12.2.1.2', '.1.3.6.1.4.1.50224.3.12.3.1.5', '.1.3.6.1.4.1.50224.3.12.3.1.4', '.1.3.6.1.4.1.50224.3.12.2.1.5'),
('zte', '.1.3.6.1.4.1.3902.1082.500.10.2.3.3.1.2', '.1.3.6.1.4.1.3902.1082.3.50.12.1.1.14', '.1.3.6.1.4.1.3902.1082.500.20.2.2.2.1.10', '.1.3.6.1.4.1.3902.1082.500.10.2.3.8.1.4');`

	result := database.DB.Exec(sql)
	if result.Error != nil {
		log.Fatalf("Failed to insert mappings: %v", result.Error)
	}

	log.Printf("Inserted OID mappings successfully. Rows affected: %d", result.RowsAffected)
}
