package utils

import (
	"net"
	"time"
)

func CheckTCPPortOpen(addr string) bool {
	timeout := 500 * time.Millisecond
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}
