package main

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func getHTTPText(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.WithField("url", url).WithError(err).Error("fail to get response")
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithField("url", url).WithError(err).Error("fail to get response")
		return ""
	}
	return string(body)
}

func getMetadata(name string) string {
	return getHTTPText(fmt.Sprintf("http://169.254.169.254/latest/meta-data/%s", name))
}

func getEnvKeys(key ...string) (value string) {
	for _, k := range key {
		value = os.Getenv(k)
		if value != "" {
			return
		}
	}
	return
}

func getExecOutput(name string, args ...string) string {
	bin, err := exec.LookPath(name)
	if err != nil {
		log.WithError(err).Error("fail to get exec output")
	}
	cmd := exec.Command(bin, args...)
	buf := new(strings.Builder)
	err = cmd.Start()
	if err != nil {
		log.WithError(err).Error("fail to get exec output")
	}
	out, _ := cmd.StdoutPipe()
	_, _ = io.Copy(buf, out)
	_ = cmd.Wait()
	return buf.String()
}

func getZilliqaMainProcess() *process.Process {
	processes, err := process.Processes()
	if err != nil {
		log.WithError(err).Error("fail to get zilliqa main process")
	}
	for _, p := range processes {
		connections, err := p.Connections()
		if err != nil {
			continue
		}
		for _, conn := range connections {
			if (conn.Type == syscall.AF_INET || conn.Type == syscall.AF_CCITT) && conn.Laddr.Port == 33133 {
				return p
			}
		}
	}
	return nil
}
