package ddl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ddlUserMessage() (*MessageInfo, error) {
	u, err := userMessage()
	if err != nil {
		return nil, err
	}
	return NewMessageInfo(*u)
}

func TestListMessageOptions(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	opts, err := listMessageOptions(*u)
	assert.NoError(t, err)
	assert.Len(t, opts, 1)
	assert.Equal(t, MessageOption{Name: "taehoio.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}, opts[0])
}

func TestExtractFields(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	fields, err := extractFields(*u)
	assert.NoError(t, err)
	assert.Len(t, fields, 7)
}

func TestExtractKeyFields(t *testing.T) {
	u, err := ddlUserMessage()
	assert.NoError(t, err)

	fieldNames, err := extractKeyFields(u.Fields)
	assert.NoError(t, err)
	assert.Len(t, fieldNames, 1)
}

func TestExtractIndices(t *testing.T) {
	u, err := ddlUserMessage()
	assert.NoError(t, err)

	indices, err := extractIndices(u.Fields)
	assert.NoError(t, err)
	assert.Len(t, indices, 1)
}

func TestGenerateDDLSQL(t *testing.T) {
	expectedLines := []string{
		"\nCREATE TABLE `user` (",
		"	`id` BIGINT UNSIGNED,",
		"	`created_at` TIMESTAMP(6) NULL DEFAULT NULL,",
		"	`updated_at` TIMESTAMP(6) NULL DEFAULT NULL,",
		"	`deleted_at` TIMESTAMP(6) NULL DEFAULT NULL,",
		"	`password_hash` VARCHAR(255),",
		"	`full_name` VARCHAR(255),",
		"	`email` VARCHAR(255),",
		"	PRIMARY KEY (`id`)",
		");",
		"",
		"CREATE INDEX `idx_email` ON `user` (`email`);",
	}
	expected := strings.Join(expectedLines, "\n")

	u, err := ddlUserMessage()
	assert.NoError(t, err)

	s, err := u.GenerateDDL()
	assert.NoError(t, err)
	assert.Equal(t, expected, s)
}

func TestGenerateDDLSQL_FailsWithSQLLite(t *testing.T) {
	u, err := ddlUserMessage()
	assert.NoError(t, err)
	u.MessageOptions = []MessageOption{{Name: "taehoio.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_SQLITE"}}

	s, err := u.GenerateDDL()
	assert.Error(t, ErrNotSupportedDatastore, err)
	assert.Empty(t, s)
}
