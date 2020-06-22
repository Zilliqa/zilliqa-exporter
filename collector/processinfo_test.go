package collector

import (
	asserting "github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var cmdline = strings.Split("zilliqa --privk 9C980BB882C69FD25149EF687A03038D1A47F4AE2F52E7D1CB35DD054E800071 --pubk 03506F19E90C97B8222A56300BE61E8EC5CAF68319B5357D6BE85C4657A1416B05 --address 34.213.52.223 --port 33133 --synctype 0 --logpath /run/zilliqa/", " ")

func TestCmdlineProcess(t *testing.T) {
	assert := asserting.New(t)
	typ, err := GetSyncTypeFromCmdline(cmdline)
	assert.NoError(err)
	assert.Equal(typ, 0)

	idx, err := GetNodeIndexFromCmdline(cmdline)
	assert.Error(err)
	assert.Equal(idx, 0)

	nt := GetNodeTypeFromCmdline(cmdline)
	assert.Equal(nt, "")
}
