package main

import (
	asserting "github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func createTestDB() string {
	tmp, err := ioutil.TempDir("", "test-db*")
	if err != nil {
		panic(err)
	}
	db, err := leveldb.OpenFile(tmp, &opt.Options{ErrorIfExist: true})
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db.Put([]byte("test"), []byte("test"), nil)
	return tmp
}

func TestCheckCorruption(t *testing.T) {
	assert := asserting.New(t)
	path := createTestDB()
	os.Remove(filepath.Join(path, "CURRENT"))
	t.Log(checkCorruption(path))
	assert.Error(checkCorruption(path))
}
