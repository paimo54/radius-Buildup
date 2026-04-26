package olt

import (
	"fmt"
	"strings"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gosnmp/gosnmp"
)

type OnuData struct {
	Index  string  `json:"index"`
	Name   string  `json:"name"`
	Status string  `json:"status"`
	Tx     float64 `json:"tx"`
	Rx     float64 `json:"rx"`
}

// ScanONU connects to an OLT via SNMP and retrieves all ONUs based on vendor OID mappings.
func ScanONU(oltID string) ([]OnuData, error) {
	var olt models.Olt
	if err := database.DB.First(&olt, "id = ?", oltID).Error; err != nil {
		return nil, fmt.Errorf("olt not found: %w", err)
	}

	var mapping models.OidMapping
	if err := database.DB.Where("vendor = ?", olt.Vendor).First(&mapping).Error; err != nil {
		return nil, fmt.Errorf("oid mapping not found for vendor %s: %w", olt.Vendor, err)
	}

	// Configure SNMP Client
	snmpClient := &gosnmp.GoSNMP{
		Target:    olt.Host,
		Port:      uint16(olt.Port),
		Community: olt.Community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(5) * time.Second,
		Retries:   2,
	}

	if err := snmpClient.Connect(); err != nil {
		return nil, fmt.Errorf("snmp connection failed: %w", err)
	}
	defer snmpClient.Conn.Close()

	// Maps to hold data by ONU index
	names := make(map[string]string)
	statuses := make(map[string]string)
	txs := make(map[string]float64)
	rxs := make(map[string]float64)

	// Helper function to extract the index suffix from a full OID
	// e.g., base OID is .1.3.6... and returned OID is .1.3.6...1234
	getIndex := func(fullOid, baseOid string) string {
		// Clean up leading dots
		full := strings.TrimPrefix(fullOid, ".")
		base := strings.TrimPrefix(baseOid, ".")
		idx := strings.TrimPrefix(full, base)
		return strings.TrimPrefix(idx, ".")
	}

	// 1. Walk OidName
	err := snmpClient.Walk(mapping.OidName, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, mapping.OidName)
		if str, ok := pdu.Value.([]byte); ok {
			names[idx] = string(str)
		} else if str, ok := pdu.Value.(string); ok {
			names[idx] = str
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk OidName: %w", err)
	}

	// 2. Walk OidStatus
	_ = snmpClient.Walk(mapping.OidStatus, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, mapping.OidStatus)
		if val, ok := pdu.Value.(int); ok {
			statuses[idx] = fmt.Sprintf("%d", val)
		} else if str, ok := pdu.Value.([]byte); ok {
			statuses[idx] = string(str)
		} else {
			statuses[idx] = fmt.Sprintf("%v", pdu.Value)
		}
		return nil
	})

	// 3. Walk OidTx
	_ = snmpClient.Walk(mapping.OidTx, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, mapping.OidTx)
		if val, ok := pdu.Value.(int); ok {
			// Typical optical power is integer / 10.0 or 100.0 (Vendor specific, assuming 10.0 for now)
			txs[idx] = float64(val) / 10.0
		}
		return nil
	})

	// 4. Walk OidRx
	_ = snmpClient.Walk(mapping.OidRx, func(pdu gosnmp.SnmpPDU) error {
		idx := getIndex(pdu.Name, mapping.OidRx)
		if val, ok := pdu.Value.(int); ok {
			rxs[idx] = float64(val) / 10.0
		}
		return nil
	})

	// Compile results
	var results []OnuData
	for idx, name := range names {
		// Clean the string (sometimes contains null bytes or unprintable chars)
		cleanName := strings.ToValidUTF8(name, "")
		
		status := "Unknown"
		if s, exists := statuses[idx]; exists {
			status = s
			// Common status mapping (1=online, 2=offline, etc. - adjust per vendor)
			if s == "1" || strings.ToLower(s) == "online" || strings.ToLower(s) == "up" {
				status = "Online"
			} else if s == "2" || s == "0" || strings.ToLower(s) == "offline" || strings.ToLower(s) == "down" {
				status = "Offline"
			}
		}

		results = append(results, OnuData{
			Index:  idx,
			Name:   cleanName,
			Status: status,
			Tx:     txs[idx],
			Rx:     rxs[idx],
		})
	}

	return results, nil
}
