package collector

import (
	asserting "github.com/stretchr/testify/assert"
	"testing"
)

func TestDetectNodeTypeFromPOD_NAME(t *testing.T) {
	assert := asserting.New(t)
	nt := nodeTypeFromPodName("xxx-xxx-seed-apipub-0")
	assert.Equal(SeedAPIPub, nt)
	assert.Equal(SeedAPIPub.String(), nt.String())
}
