package main

import (
	"fmt"
	"log"

	"radius-buildup/internal/config"
	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	_, err = database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Cek apakah admin sudah ada
	var admin models.AdminUser
	result := database.DB.Where("username = ?", "superadmin").First(&admin)

	if result.Error != nil {
		// Buat admin baru
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		
		admin = models.AdminUser{
			ID:       "admin_1",
			Username: "superadmin",
			Password: string(hashedPassword),
			Name:     "Super Admin",
			Role:     "SUPER_ADMIN",
			IsActive: true,
		}

		if err := database.DB.Create(&admin).Error; err != nil {
			log.Fatalf("Failed to create admin: %v", err)
		}
		fmt.Println("✅ Admin 'superadmin' berhasil dibuat (Password: admin123)")
	} else {
		// Update password ke admin123 jika sudah ada (agar pasti tahu)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		database.DB.Model(&admin).Update("password", string(hashedPassword))
		fmt.Println("✅ Admin 'superadmin' sudah ada. Password direset ke: admin123")
	}
}
