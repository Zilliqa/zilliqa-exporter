package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/zilliqa/zilliqa-exporter/utils"
)

func checkCorruption(path string) error {
	if !utils.PathIsDir(path) {
		log.Fatalf("path %s not exists or is not a dir", path)
	}
	db, err := leveldb.OpenFile(path, &opt.Options{ReadOnly: true, ErrorIfMissing: true})
	if err != nil {
		return err
	}
	err = db.Close()
	return err
}
