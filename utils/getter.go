package utils

import (
	"context"
	"fmt"
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
	return CheckTCPPortOpen(AWSMetadataIP+":80", 100 * time.Millisecond) == nil
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
	bin = name
	if name == "" {
		log.Error("GetExecOutput: executable name should not be empty")
		return ""
	}
	if !PathExists(name) {
		bin, err = exec.LookPath(name)
		if err != nil {
			log.WithError(err).Error("fail to get exec output")
		}
	}
	cmd := exec.Command(bin, args...)
	log.Debugf("exec cmd: %s %s", bin, strings.Join(args, " "))
	buf := new(strings.Builder)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.WithError(err).Error("fail to get process stdout pipe")
	}
	err = cmd.Start()
	if err != nil {
		log.WithError(err).Error("fail to get exec output")
	}
	_, err = io.Copy(buf, out)
	_ = cmd.Wait()
	return buf.String()
}
