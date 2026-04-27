package olt

import (
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gosnmp/gosnmp"
)

type OnuData struct {
	Index    string  `json:"index"`
	Name     string  `json:"name"`
	Serial   string  `json:"serial"`
	Model    string  `json:"model"`
	Status   string  `json:"status"`
	Tx       float64 `json:"tx"`
	Rx       float64 `json:"rx"`
	Distance float64 `json:"distance"`
}

// getIndex extracts the index suffix from a full OID relative to a base OID.
// e.g., full=".1.3.6.1.4.1.3902.1012.3.28.1.1.2.268501248.1" base=".1.3.6.1.4.1.3902.1012.3.28.1.1.2"
// returns "268501248.1"
func getIndex(fullOid, baseOid string) string {
	full := strings.TrimPrefix(fullOid, ".")
	base := strings.TrimPrefix(baseOid, ".")
	idx := strings.TrimPrefix(full, base)
	return strings.TrimPrefix(idx, ".")
}

// createSNMPClient creates a new GoSNMP client for the given OLT.
// Each goroutine needs its own client since gosnmp is not goroutine-safe.
func createSNMPClient(host string, port uint16, community string) (*gosnmp.GoSNMP, error) {
	client := &gosnmp.GoSNMP{
		Target:         host,
		Port:           port,
		Community:      community,
		Version:        gosnmp.Version2c,
		Timeout:        time.Duration(10) * time.Second,
		Retries:        1,
		MaxRepetitions: 50, // BulkWalk: fetch 50 rows per packet (default is 10)
	}
	if err := client.Connect(); err != nil {
		return nil, err
	}
	return client, nil
}

// bulkWalkStrings performs a BulkWalk for string-type OIDs (name, serial, model)
func bulkWalkStrings(client *gosnmp.GoSNMP, oid string) (map[string]string, error) {
	result := make(map[string]string)
	if oid == "" {
		return result, nil
	}
	err := client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, oid)
		switch v := pdu.Value.(type) {
		case []byte:
			result[idx] = strings.TrimSpace(strings.ToValidUTF8(string(v), ""))
		case string:
			result[idx] = strings.TrimSpace(v)
		case int:
			result[idx] = fmt.Sprintf("%d", v)
		}
		return nil
	})
	return result, err
}

// formatSerial converts raw SN bytes to standard GPON format (e.g. ZTEG1234ABCD)
func formatSerial(b []byte) string {
	if len(b) >= 8 {
		// GPON SN usually has 4 ASCII chars vendor ID + 4 hex bytes
		vendor := string(b[:4])
		isAlpha := true
		for _, c := range vendor {
			if c < 32 || c > 126 {
				isAlpha = false
				break
			}
		}
		if isAlpha {
			return fmt.Sprintf("%s%X", vendor, b[4:])
		}
	}
	// Fallback to all hex if vendor is not ASCII printable or length is short
	return fmt.Sprintf("%X", b)
}

// bulkWalkSerial performs a BulkWalk specifically for Serial Numbers to handle hex formatting
func bulkWalkSerial(client *gosnmp.GoSNMP, oid string) (map[string]string, error) {
	result := make(map[string]string)
	if oid == "" {
		return result, nil
	}
	err := client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, oid)
		switch v := pdu.Value.(type) {
		case []byte:
			result[idx] = formatSerial(v)
		case string:
			// If it's already a string but might contain unprintable chars
			result[idx] = formatSerial([]byte(v))
		}
		return nil
	})
	return result, err
}

// bulkWalkInts performs a BulkWalk for integer-type OIDs (status)
func bulkWalkInts(client *gosnmp.GoSNMP, oid string) (map[string]string, error) {
	result := make(map[string]string)
	if oid == "" {
		return result, nil
	}
	err := client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, oid)
		switch v := pdu.Value.(type) {
		case int:
			result[idx] = fmt.Sprintf("%d", v)
		case []byte:
			result[idx] = string(v)
		default:
			result[idx] = fmt.Sprintf("%v", pdu.Value)
		}
		return nil
	})
	return result, err
}

// bulkWalkFloats performs a BulkWalk for float-type OIDs (tx, rx power) with divider
func bulkWalkFloats(client *gosnmp.GoSNMP, oid string, divider float64) (map[string]float64, error) {
	result := make(map[string]float64)
	if oid == "" {
		return result, nil
	}
	err := client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, oid)
		if val, ok := pdu.Value.(int); ok {
			result[idx] = float64(val) / divider
		}
		return nil
	})
	return result, err
}

// bulkWalkRawInts performs a BulkWalk for raw integer values (distance in meters)
func bulkWalkRawInts(client *gosnmp.GoSNMP, oid string) (map[string]float64, error) {
	result := make(map[string]float64)
	if oid == "" {
		return result, nil
	}
	err := client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, oid)
		if val, ok := pdu.Value.(int); ok {
			result[idx] = float64(val)
		}
		return nil
	})
	return result, err
}

// ScanONU connects to an OLT via SNMP and retrieves all ONUs based on vendor OID mappings.
// Uses BulkWalk + parallel goroutines for maximum speed.
func ScanONU(oltID string) ([]OnuData, error) {
	start := time.Now()

	var olt models.Olt
	if err := database.DB.First(&olt, "id = ?", oltID).Error; err != nil {
		return nil, fmt.Errorf("olt not found: %w", err)
	}

	var mapping models.OidMapping
	if err := database.DB.Where("vendor = ?", olt.Vendor).First(&mapping).Error; err != nil {
		return nil, fmt.Errorf("oid mapping not found for vendor %s: %w", olt.Vendor, err)
	}

	port := uint16(olt.Port)
	if port == 0 {
		port = 161
	}
	community := olt.Community
	if community == "" {
		community = "public"
	}

	log.Printf("[SNMP] Scan OLT %s (%s:%d) vendor=%s", olt.Name, olt.Host, port, olt.Vendor)

	// Quick connectivity test
	testClient, err := createSNMPClient(olt.Host, port, community)
	if err != nil {
		return nil, fmt.Errorf("snmp connection failed: %w", err)
	}

	testResult, err := testClient.Get([]string{".1.3.6.1.2.1.1.1.0"})
	testClient.Conn.Close()
	if err != nil {
		log.Printf("[SNMP ERROR] OLT %s tidak merespons SNMP di %s:%d", olt.Name, olt.Host, port)
		return nil, fmt.Errorf("OLT tidak merespons SNMP di %s:%d (community: %s). Pastikan SNMP aktif, port benar, dan firewall terbuka.", olt.Host, port, community)
	}
	if len(testResult.Variables) > 0 {
		if val, ok := testResult.Variables[0].Value.([]byte); ok {
			log.Printf("[SNMP] sysDescr: %s", string(val))
		}
	}

	// =========================================================================
	// Parallel BulkWalk — each OID tree gets its own goroutine + SNMP connection
	// =========================================================================
	var (
		mu        sync.Mutex
		wg        sync.WaitGroup
		names     map[string]string
		serials   map[string]string
		modelMap  map[string]string
		statuses  map[string]string
		txs       map[string]float64
		rxs       map[string]float64
		distances map[string]float64
		firstErr  error
	)

	// Define all walk tasks
	type walkTask struct {
		name string
		oid  string
	}

	stringTasks := []walkTask{
		{"Name", mapping.OidName},
		{"Serial", mapping.OidSerial},
		{"Model", mapping.OidModel},
	}

	stringResults := make([]map[string]string, len(stringTasks))

	// Launch string walks in parallel
	for i, task := range stringTasks {
		if task.oid == "" {
			stringResults[i] = make(map[string]string)
			continue
		}
		wg.Add(1)
		go func(idx int, t walkTask) {
			defer wg.Done()
			client, err := createSNMPClient(olt.Host, port, community)
			if err != nil {
				log.Printf("[SNMP ERROR] Failed to connect for %s: %v", t.name, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			defer client.Conn.Close()

			var result map[string]string
			if t.name == "Serial" {
				result, err = bulkWalkSerial(client, t.oid)
			} else {
				result, err = bulkWalkStrings(client, t.oid)
			}
			if err != nil {
				log.Printf("[SNMP WARN] BulkWalk %s failed: %v", t.name, err)
			}
			log.Printf("[SNMP] %s: %d entries", t.name, len(result))
			stringResults[idx] = result
		}(i, task)
	}

	// Launch status walk
	wg.Add(1)
	go func() {
		defer wg.Done()
		client, err := createSNMPClient(olt.Host, port, community)
		if err != nil {
			return
		}
		defer client.Conn.Close()
		result, err := bulkWalkInts(client, mapping.OidStatus)
		if err != nil {
			log.Printf("[SNMP WARN] BulkWalk Status failed: %v", err)
		}
		log.Printf("[SNMP] Status: %d entries", len(result))
		mu.Lock()
		statuses = result
		mu.Unlock()
	}()

	// Launch Tx walk
	wg.Add(1)
	go func() {
		defer wg.Done()
		client, err := createSNMPClient(olt.Host, port, community)
		if err != nil {
			return
		}
		defer client.Conn.Close()
		
		// Custom handler for Tx power since ZTE returns it in 0.01 uW when it's positive
		result := make(map[string]float64)
		if mapping.OidTx != "" {
			err = client.BulkWalk(mapping.OidTx, func(pdu gosnmp.SnmpPDU) error {
				idx := getIndex(pdu.Name, mapping.OidTx)
				if val, ok := pdu.Value.(int); ok {
					if val < 0 {
						// Standard 0.1 dBm format (e.g., -250 = -25.0 dBm)
						result[idx] = float64(val) / 10.0
					} else if val > 0 {
						// ZTE format: 0.01 uW. Convert to dBm: 10 * log10( uW / 1000 )
						// val / 100.0 = uW -> uW / 1000.0 = mW -> 10 * log10(val / 100000.0)
						result[idx] = 10 * math.Log10(float64(val)/100000.0)
					}
				}
				return nil
			})
			if err != nil {
				log.Printf("[SNMP WARN] BulkWalk Tx failed: %v", err)
			}
		}
		
		log.Printf("[SNMP] Tx: %d entries", len(result))
		mu.Lock()
		txs = result
		mu.Unlock()
	}()

	// Launch Rx walk
	wg.Add(1)
	go func() {
		defer wg.Done()
		client, err := createSNMPClient(olt.Host, port, community)
		if err != nil {
			return
		}
		defer client.Conn.Close()
		result, err := bulkWalkFloats(client, mapping.OidRx, 10.0)
		if err != nil {
			log.Printf("[SNMP WARN] BulkWalk Rx failed: %v", err)
		}
		log.Printf("[SNMP] Rx: %d entries", len(result))
		mu.Lock()
		rxs = result
		mu.Unlock()
	}()

	// Launch Distance walk (if configured)
	if mapping.OidDistance != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := createSNMPClient(olt.Host, port, community)
			if err != nil {
				return
			}
			defer client.Conn.Close()
			result, err := bulkWalkRawInts(client, mapping.OidDistance)
			if err != nil {
				log.Printf("[SNMP WARN] BulkWalk Distance failed: %v", err)
			}
			log.Printf("[SNMP] Distance: %d entries", len(result))
			mu.Lock()
			distances = result
			mu.Unlock()
		}()
	} else {
		distances = make(map[string]float64)
	}

	// Wait for all walks to complete
	wg.Wait()

	// Assign string results
	names = stringResults[0]
	serials = stringResults[1]
	modelMap = stringResults[2]

	if names == nil {
		names = make(map[string]string)
	}
	if serials == nil {
		serials = make(map[string]string)
	}
	if modelMap == nil {
		modelMap = make(map[string]string)
	}
	if statuses == nil {
		statuses = make(map[string]string)
	}
	if txs == nil {
		txs = make(map[string]float64)
	}
	if rxs == nil {
		rxs = make(map[string]float64)
	}

	// Check if we got any data at all
	if len(names) == 0 {
		if firstErr != nil {
			return nil, fmt.Errorf("SNMP scan gagal: %w", firstErr)
		}
		log.Printf("[SNMP] No ONU data found. OID mungkin tidak cocok dengan perangkat ini.")
		return []OnuData{}, nil
	}

	// Debug: compare index formats between different OID trees
	sampleCount := 3
	log.Printf("[SNMP DEBUG] === Index Format Comparison ===")
	i := 0
	for idx := range names {
		if i >= sampleCount { break }
		log.Printf("[SNMP DEBUG] Name idx: '%s' → has tx=%v rx=%v dist=%v status=%v",
			idx, txs[idx] != 0, rxs[idx] != 0, distances[idx] != 0, statuses[idx] != "")
		i++
	}
	i = 0
	for idx := range txs {
		if i >= sampleCount { break }
		log.Printf("[SNMP DEBUG] Tx idx: '%s' → value=%.2f", idx, txs[idx])
		i++
	}
	i = 0
	for idx := range rxs {
		if i >= sampleCount { break }
		log.Printf("[SNMP DEBUG] Rx idx: '%s' → value=%.2f", idx, rxs[idx])
		i++
	}
	i = 0
	for idx := range distances {
		if i >= sampleCount { break }
		log.Printf("[SNMP DEBUG] Distance idx: '%s' → value=%.0f", idx, distances[idx])
		i++
	}

	// Compile results
	var results []OnuData
	// Helper to look up float values with fallback to ".1" suffix (ZTE C320 uses PON.ONU.1 for optical data)
	lookupFloat := func(m map[string]float64, idx string) float64 {
		if v, ok := m[idx]; ok && v != 0 {
			return v
		}
		// Try with ".1" suffix (ZTE optical power OID index format)
		if v, ok := m[idx+".1"]; ok {
			return v
		}
		return 0
	}

	for idx, name := range names {
		cleanName := strings.ToValidUTF8(name, "")

		status := "Unknown"
		if s, exists := statuses[idx]; exists {
			status = s
			// ZTE status mapping: 1=online(logging), 2=offline, 3=online(active)
			switch s {
			case "1", "3":
				status = "Online"
			case "2", "0":
				status = "Offline"
			default:
				lower := strings.ToLower(s)
				if lower == "online" || lower == "up" {
					status = "Online"
				} else if lower == "offline" || lower == "down" {
					status = "Offline"
				}
			}
		}

		results = append(results, OnuData{
			Index:    idx,
			Name:     cleanName,
			Serial:   serials[idx],
			Model:    modelMap[idx],
			Status:   status,
			Tx:       lookupFloat(txs, idx),
			Rx:       lookupFloat(rxs, idx),
			Distance: lookupFloat(distances, idx),
		})
	}

	elapsed := time.Since(start)
	log.Printf("[SNMP] Scan complete: %d ONUs in %v", len(results), elapsed)

	return results, nil
}
