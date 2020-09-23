package collector

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

func GetZilliqaMainProcess(constants *Constants) *process.Process {
	procs, err := GetProcesses("zilliqa", constants.P2PPort(), 4201, 4301)
	if err != nil {
		log.WithError(err).Error("fail to get zilliqa main process")
		return nil
	}
	if len(procs) > 0 {
		proc := procs[0]
		name, err := proc.Name()
		if err != nil {
			log.WithError(err).Error("fail to get zilliqa main process")
			return nil
		} else if strings.EqualFold(filepath.Base(name), "scilla-server") {
			log.Warn("scilla-server may inherited the p2p port of zilliqa process")
			log.Error("fail to get zilliqa main process")
			return nil
		}
		return proc
	}
	return nil
}

func GetZilliqadProcess() *process.Process {
	processes, err := process.Processes()
	if err != nil {
		log.WithError(err).Error("fail to get zilliqad main process")
		return nil
	}
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			log.WithError(err).Error("fail to get zilliqad main process")
			return nil
		}
		if filepath.Base(name) == "zilliqad" {
			return proc
		}
	}
	return nil
}

// match port first, if no matched, return procs that matches name
func GetProcesses(name string, port ...uint32) ([]*process.Process, error) {
	processes, err := process.Processes()
	if name == "" && len(port) == 0 {
		return processes, err
	}
	if err != nil {
		return nil, err
	}
	var portMatched []*process.Process
	var nameMatched []*process.Process
Loop:
	for _, proc := range processes {
		connections, err := proc.Connections()
		if err != nil {
			continue
		}
		for _, p := range port {
			for _, conn := range connections {
				if conn.Laddr.Port == p {
					portMatched = append(portMatched, proc)
					continue Loop
				}
			}
		}

		name, err := proc.Name()
		if err != nil {
			return nil, err
		}
		if filepath.Base(name) == name {
			nameMatched = append(nameMatched, proc)
		}
	}
	if len(port) > 0 && len(portMatched) > 0 {
		return portMatched, nil
	}
	if name != "" && len(nameMatched) > 0 {
		return nameMatched, nil
	}
	return nil, errors.New("process not found")
}
