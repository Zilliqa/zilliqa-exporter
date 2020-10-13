package utils

import (
	"github.com/pkg/errors"
	"net"
	"time"
)

func CheckTCPPortOpen(addr string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return errors.Wrap(err, "check tcp port fail")
	}
	if conn != nil {
		defer conn.Close()
		return nil
	}
	return errors.New("check tcp port fail: connection & error both nil")
}
