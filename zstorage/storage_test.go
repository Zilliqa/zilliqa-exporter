package zstorage

import (
	"github.com/gogo/protobuf/jsonpb"
	asserting "github.com/stretchr/testify/assert"
	"github.com/zilliqa/zilliqa-exporter/zstorage/ZilliqaMessage"
	"testing"
)

func TestStorage(t *testing.T) {
	assert := asserting.New(t)
	db := NewReadOnlyDB("persistence/VCBlocks")
	defer db.Close()
	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	assert.True(iter.Next())
	t.Log(iter.Key(), iter.Value())
	model := &ZilliqaMessage.ProtoVCBlock{}
	assert.NoError(model.Unmarshal(iter.Value()))
	t.Log(model.String())
	m := &jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "  ",
		OrigName:     true,
		AnyResolver:  nil,
	}
	t.Log(m.MarshalToString(model))
	t.Log(db.GetProperty("leveldb.sstables"))
	t.Log(db.GetProperty("leveldb.stats"))
}
