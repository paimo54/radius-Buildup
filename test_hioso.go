package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gosnmp/gosnmp"
)

func main() {
	client := &gosnmp.GoSNMP{
		Target:    "192.168.0.101",
		Port:      161,
		Community: "public",
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(5) * time.Second,
		Retries:   2,
		MaxOids:   gosnmp.MaxOids,
	}

	err := client.Connect()
	if err != nil {
		log.Fatalf("Connect err: %v", err)
	}
	defer client.Conn.Close()

	// Walk Name
	fmt.Println("--- WALKING NAME ---")
	count := 0
	_ = client.BulkWalk(".1.3.6.1.4.1.25355.3.2.6.3.2.1.37", func(pdu gosnmp.SnmpPDU) error {
		if count < 5 {
			fmt.Printf("Name OID: %s, Value: %v\n", pdu.Name, pdu.Value)
		}
		count++
		return nil
	})
	fmt.Printf("Total Name entries: %d\n\n", count)

	// Walk Tx
	fmt.Println("--- WALKING TX ---")
	count = 0
	_ = client.BulkWalk(".1.3.6.1.4.1.25355.3.2.6.14.2.1.4", func(pdu gosnmp.SnmpPDU) error {
		if count < 5 {
			fmt.Printf("Tx OID: %s, Value: %v\n", pdu.Name, pdu.Value)
		}
		count++
		return nil
	})
	fmt.Printf("Total Tx entries: %d\n\n", count)
}
