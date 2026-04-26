package mikrotik

import (
	"fmt"
	"strings"

	"radius-buildup/internal/models"

	"github.com/go-routeros/routeros"
)

// DialRouter connects to a MikroTik router using the API
func DialRouter(router *models.Router) (*routeros.Client, error) {
	addr := fmt.Sprintf("%s:%d", router.IPAddress, router.ApiPort)
	
	client, err := routeros.Dial(addr, router.Username, router.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return client, nil
}

// PingRouter checks if the router is reachable via API
func PingRouter(router *models.Router) (bool, string) {
	client, err := DialRouter(router)
	if err != nil {
		return false, err.Error()
	}
	defer client.Close()

	// Try a simple command
	_, err = client.Run("/system/resource/print")
	if err != nil {
		return false, err.Error()
	}

	return true, "OK"
}

// KickPppoeUser disconnects an active PPPoE session by username
func KickPppoeUser(router *models.Router, username string) error {
	client, err := DialRouter(router)
	if err != nil {
		return err
	}
	defer client.Close()

	// 1. Find the active session
	reply, err := client.Run("/ppp/active/print", "?name="+username)
	if err != nil {
		return fmt.Errorf("failed to query active sessions: %w", err)
	}

	if len(reply.Re) == 0 {
		// Not found, maybe already disconnected
		return nil
	}

	// 2. Remove the session
	var errs []string
	for _, re := range reply.Re {
		id := re.Map[".id"]
		if id != "" {
			_, err = client.Run("/ppp/active/remove", "=.id="+id)
			if err != nil {
				errs = append(errs, err.Error())
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during remove: %s", strings.Join(errs, ", "))
	}

	return nil
}

// KickHotspotUser disconnects an active Hotspot session by username
func KickHotspotUser(router *models.Router, username string) error {
	client, err := DialRouter(router)
	if err != nil {
		return err
	}
	defer client.Close()

	// 1. Find the active session
	reply, err := client.Run("/ip/hotspot/active/print", "?user="+username)
	if err != nil {
		return fmt.Errorf("failed to query active hotspot: %w", err)
	}

	if len(reply.Re) == 0 {
		return nil
	}

	// 2. Remove the session
	var errs []string
	for _, re := range reply.Re {
		id := re.Map[".id"]
		if id != "" {
			_, err = client.Run("/ip/hotspot/active/remove", "=.id="+id)
			if err != nil {
				errs = append(errs, err.Error())
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during remove: %s", strings.Join(errs, ", "))
	}

	return nil
}

// GetSystemResource returns CPU/RAM usage of the router
func GetSystemResource(router *models.Router) (map[string]string, error) {
	client, err := DialRouter(router)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.Run("/system/resource/print")
	if err != nil {
		return nil, err
	}

	if len(reply.Re) > 0 {
		return reply.Re[0].Map, nil
	}

	return nil, fmt.Errorf("no resource data returned")
}
