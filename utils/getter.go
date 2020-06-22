package utils

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const AWSMetadataIP = "169.254.169.254"
const AWSMetadataUrl = "http://169.254.169.254/latest/meta-data"

func GetHTTPText(url string) string {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.WithField("url", url).WithError(err).Error("fail to build request")
		return ""
	}
	resp, err := http.DefaultClient.Do(req)
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

func MetadataAvailable() bool {
	return CheckTCPPortOpen(AWSMetadataIP + ":80")
}

func GetMetadata(name string) string {
	return GetHTTPText(fmt.Sprintf("%s/%s", AWSMetadataUrl, name))
}

func GetEnvKeys(key ...string) (value string) {
	for _, k := range key {
		value = os.Getenv(k)
		if value != "" {
			return
		}
	}
	return
}

func GetExecOutput(name string, args ...string) string {
	var bin string
	var err error
	if !PathExists(name) {
		bin, err = exec.LookPath(name)
		if err != nil {
			log.WithError(err).Error("fail to get exec output")
		}
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

func GetZilliqaMainProcess() *process.Process {
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
			if conn.Laddr.Port == 33133 {
				return p
			}
		}
	}
	return nil
}
