package ddl

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	version                 = "0.0.1-alpha"
	GeneratedFilenamePrefix = "github.com/taehoio/protoc-gen-go-ddl/gen/go/"
)

func TestGenerateDDLFiles(t *testing.T) {
	gen, err := gen()
	assert.NoError(t, err)

	err = GenerateDDLFiles(version, gen)
	for _, generatedFile := range gen.Response().File {
		assert.NotNil(t, generatedFile)

		goldenFilePath := fmt.Sprintf("../../testdata/gen/go/taehoio/%s", strings.TrimPrefix(*generatedFile.Name, GeneratedFilenamePrefix))
		goldenFileContent, err := os.ReadFile(goldenFilePath)
		assert.NoError(t, err)

		assert.Equal(t, string(goldenFileContent), *generatedFile.Content)
	}
	assert.NoError(t, err)
}
