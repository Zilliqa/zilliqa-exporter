package collector

import (
	asserting "github.com/stretchr/testify/assert"
	"testing"
)

func TestDetectNodeTypeFromPOD_NAME(t *testing.T) {
	assert := asserting.New(t)
	nt, idx := nodeTypeIndexFromPodName("xxx-xxx-seed-apipub-0")
	assert.Equal(idx, 0)
	assert.Equal(SeedPub, nt)
	assert.Equal(SeedPub.String(), nt.String())
}
