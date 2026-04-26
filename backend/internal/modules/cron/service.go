package cron

import (
	"log"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"
	"radius-buildup/internal/modules/pppoe"

	"github.com/robfig/cron/v3"
)

var cronRunner *cron.Cron

// InitCron starts all background cron jobs
func InitCron() {
	// Use Asia/Jakarta timezone for cron jobs
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		location = time.Local
	}

	cronRunner = cron.New(cron.WithLocation(location))

	// 1. PPPoE Auto Isolir - Every hour at minute 0
	cronRunner.AddFunc("0 * * * *", RunAutoIsolir)

	// 2. PPPoE Session Sync - Every 5 minutes
	// cronRunner.AddFunc("*/5 * * * *", RunSessionSync)

	// Start the cron scheduler asynchronously
	cronRunner.Start()
	log.Println("⏱️ Cron scheduler started successfully")

	// Run AutoIsolir once on startup (with 10s delay to allow DB init)
	go func() {
		time.Sleep(10 * time.Second)
		log.Println("[CRON] Running AutoIsolir on startup...")
		RunAutoIsolir()
	}()
}

// StopCron gracefully stops the scheduler
func StopCron() {
	if cronRunner != nil {
		cronRunner.Stop()
	}
}

// RunAutoIsolir checks for active users whose expiration date has passed and isolates them.
func RunAutoIsolir() {
	now := time.Now()
	log.Println("[CRON] Starting PPPoE Auto Isolir process...")

	var users []models.PppoeUser
	// Find users that are ACTIVE but their expiredAt is in the past
	if err := database.DB.Preload("Profile").Where("status = ? AND expiredAt <= ?", "active", now).Find(&users).Error; err != nil {
		log.Printf("[CRON ERROR] Failed to fetch expired users: %v", err)
		return
	}

	if len(users) == 0 {
		log.Println("[CRON] No users to isolate.")
		return
	}

	count := 0
	for _, user := range users {
		user.Status = "isolated"
		if err := database.DB.Save(&user).Error; err != nil {
			log.Printf("[CRON ERROR] Failed to isolate user %s: %v", user.Username, err)
			continue
		}

		// Sync isolation status to FreeRADIUS
		if err := pppoe.SyncToRadius(&user); err != nil {
			log.Printf("[CRON ERROR] Failed to sync radius for user %s: %v", user.Username, err)
			// Optional: we might want to log this to activity_logs or a dedicated cron log table
		} else {
			count++
		}
	}

	log.Printf("[CRON] Auto Isolir completed. Isolated %d/%d users.", count, len(users))
}
