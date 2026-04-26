package pppoe

import (
	"errors"
	"log"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"gorm.io/gorm"
)

// SyncToRadius synchronizes a PPPoE user to FreeRADIUS tables (radcheck, radreply, radusergroup)
func SyncToRadius(user *models.PppoeUser) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Cleanup existing records for this username
		tx.Where("username = ?", user.Username).Delete(&models.Radcheck{})
		tx.Where("username = ?", user.Username).Delete(&models.Radreply{})
		tx.Where("username = ?", user.Username).Delete(&models.Radusergroup{})

		// 2. If status is 'blocked' or 'stop', we are done (keep tables clean)
		if user.Status == "blocked" || user.Status == "stop" {
			return updateSyncStatus(tx, user.ID, true)
		}

		// 3. Prepare common attributes
		// Password (radcheck)
		tx.Create(&models.Radcheck{
			Username:  user.Username,
			Attribute: "Cleartext-Password",
			Op:        ":=",
			Value:     user.Password,
		})

		// 4. Handle based on status
		if user.Status == "isolated" {
			// Isolated user gets 'isolir' group
			tx.Create(&models.Radusergroup{
				Username:  user.Username,
				Groupname: "isolir",
				Priority:  1,
			})
			// Isolated users usually don't get static IP (uses pool)
		} else {
			// Active user (default)
			// Get profile to find group name
			var profile models.PppoeProfile
			if err := tx.First(&profile, "id = ?", user.ProfileID).Error; err != nil {
				return errors.New("Profile not found for sync")
			}

			// Group (radusergroup)
			tx.Create(&models.Radusergroup{
				Username:  user.Username,
				Groupname: profile.GroupName,
				Priority:  0,
			})

			// Static IP (radreply) if assigned
			if user.IPAddress != nil && *user.IPAddress != "" {
				tx.Create(&models.Radreply{
					Username:  user.Username,
					Attribute: "Framed-IP-Address",
					Op:        ":=",
					Value:     *user.IPAddress,
				})
			}
		}

		return updateSyncStatus(tx, user.ID, true)
	})
}

// CleanupRadius removes all RADIUS records for a username
func CleanupRadius(username string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("username = ?", username).Delete(&models.Radcheck{})
		tx.Where("username = ?", username).Delete(&models.Radreply{})
		tx.Where("username = ?", username).Delete(&models.Radusergroup{})
		return nil
	})
}

func updateSyncStatus(tx *gorm.DB, userID string, status bool) error {
	now := time.Now()
	return tx.Model(&models.PppoeUser{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"syncedToRadius": status,
		"lastSyncAt":     &now,
	}).Error
}

// Helper to log errors during background sync
func LogSyncError(username string, err error) {
	log.Printf("⚠️ RADIUS Sync Error [%s]: %v", username, err)
}
