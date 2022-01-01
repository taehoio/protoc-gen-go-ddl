package ddl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListMessageOptions(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	ddlUserMsg, err := NewMessageInfo(*u)
	assert.NoError(t, err)

	opts, err := ddlUserMsg.listMessageOptions()
	assert.NoError(t, err)
	assert.Len(t, opts, 1)
	assert.Equal(t, MessageOption{Name: "ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}, opts[0])
}
