package olt

import (
	"fmt"
	"time"

	"radius-buildup/internal/models"
	"github.com/ziutek/telnet"
)

type TelnetSession struct {
	conn    *telnet.Conn
	olt     *models.Olt
	timeout time.Duration
}

// NewTelnetSession creates a new telnet session for an OLT
func NewTelnetSession(olt *models.Olt) (*TelnetSession, error) {
	addr := fmt.Sprintf("%s:%d", olt.Host, olt.TelnetPort)
	if olt.TelnetPort == 0 {
		addr = fmt.Sprintf("%s:23", olt.Host)
	}

	conn, err := telnet.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to telnet: %w", err)
	}

	return &TelnetSession{
		conn:    conn,
		olt:     olt,
		timeout: 30 * time.Second,
	}, nil
}

// Login performs the login sequence for ZTE OLT
func (s *TelnetSession) Login() error {
	s.conn.SetUnixWriteMode(true)

	// Wait for Username prompt
	if err := s.expect("Username:"); err != nil {
		return err
	}
	if err := s.send(s.olt.TelnetUser); err != nil {
		return err
	}

	// Wait for Password prompt
	if err := s.expect("Password:"); err != nil {
		return err
	}
	if err := s.send(s.olt.TelnetPass); err != nil {
		return err
	}

	// Wait for user prompt (usually ZXAN>)
	if err := s.expect(">"); err != nil {
		return fmt.Errorf("login failed or unexpected prompt: %w", err)
	}

	// Enter enable mode if password is provided
	if s.olt.TelnetEnable != "" {
		if err := s.send("enable"); err != nil {
			return err
		}
		if err := s.expect("Password:"); err != nil {
			return err
		}
		if err := s.send(s.olt.TelnetEnable); err != nil {
			return err
		}
		if err := s.expect("#"); err != nil {
			return fmt.Errorf("failed to enter enable mode: %w", err)
		}
	}

	return nil
}

// Execute sends a command and returns the output
func (s *TelnetSession) Execute(cmd string) (string, error) {
	if err := s.send(cmd); err != nil {
		return "", err
	}
	
	// Read until the next prompt
	data, err := s.conn.ReadUntil(">", "#")
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

// Close closes the telnet connection
func (s *TelnetSession) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

// Internal helpers
func (s *TelnetSession) expect(delims ...string) error {
	s.conn.SetReadDeadline(time.Now().Add(s.timeout))
	return s.conn.SkipUntil(delims...)
}

func (s *TelnetSession) send(cmd string) error {
	s.conn.SetWriteDeadline(time.Now().Add(s.timeout))
	_, err := s.conn.Write([]byte(cmd + "\n"))
	return err
}

// ============================================================================
// ZTE Specific Commands
// ============================================================================

// RebootOnu reboots a specific ONU on ZTE OLT
func (s *TelnetSession) RebootOnu(ponPort string, onuID string) (string, error) {
	// Example command for ZTE: pon-onu-mng gpon-onu_1/1/1:1
	// Then: reboot
	
	// This is a simplified version. Actual ZTE commands might require entering config mode first.
	commands := []string{
		"conf t",
		fmt.Sprintf("pon-onu-mng gpon-onu_%s:%s", ponPort, onuID),
		"reboot",
		"yes", // confirm reboot if asked
		"exit",
		"exit",
	}
	
	var lastOutput string
	for _, cmd := range commands {
		out, err := s.Execute(cmd)
		if err != nil {
			return lastOutput, err
		}
		lastOutput = out
	}
	
	return lastOutput, nil
}
