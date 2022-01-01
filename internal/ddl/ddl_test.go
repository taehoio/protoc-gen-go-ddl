package ddl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const version = "0.0.1-testing"

func TestGenerateDDLFiles(t *testing.T) {
	gen, err := gen()
	assert.NoError(t, err)

	err = GenerateDDLFiles(version, gen)
	assert.NoError(t, err)
}
