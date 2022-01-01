package ddl

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
)

func gen() (*protogen.Plugin, error) {
	// https://medium.com/@tim.r.coulson/writing-a-protoc-plugin-with-google-golang-org-protobuf-cd5aa75f5777
	b, err := ioutil.ReadFile("../../testdata/protoc_stdin.out")
	if err != nil {
		return nil, err
	}

	var genReq pluginpb.CodeGeneratorRequest
	if err := proto.Unmarshal(b, &genReq); err != nil {
		return nil, err
	}

	return protogen.Options{}.New(&genReq)
}

func userFile() (*protogen.File, error) {
	gen, err := gen()
	if err != nil {
		return nil, err
	}

	return gen.Files[3], nil
}

func userMessage() (*protogen.Message, error) {
	userFile, err := userFile()
	if err != nil {
		return nil, err
	}

	userMessage := userFile.Messages[0]

	return userMessage, nil
}

func TestKindToSQLType(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	type test struct {
		kind protoreflect.Kind
		want string
	}

	tests := []test{
		{protoreflect.Uint64Kind, "BIGINT UNSIGNED"},
		{protoreflect.MessageKind, "TIMESTAMP NULL DEFAULT NULL"},
		{protoreflect.MessageKind, "TIMESTAMP NULL DEFAULT NULL"},
		{protoreflect.MessageKind, "TIMESTAMP NULL DEFAULT NULL"},
		{protoreflect.StringKind, "VARCHAR(255)"},
		{protoreflect.StringKind, "VARCHAR(255)"},
		{protoreflect.StringKind, "VARCHAR(255)"},
	}

	for i, pbField := range u.Fields {
		ddlField, err := NewField(*pbField)
		assert.NoError(t, err)
		assert.NotEmpty(t, ddlField)

		tt := tests[i]
		assert.EqualValues(t, tt.kind, pbField.Desc.Kind())
		assert.EqualValues(t, tt.want, ddlField.Type)
	}
}

func TestListFieldOptions(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	idField, err := NewField(*u.Fields[0])
	assert.NoError(t, err)
	idOpts := idField.listFieldOptions()
	assert.Len(t, idOpts, 1)
	assert.Equal(t, FieldOption{Name: "ddl.protobuf.v1.key", Value: "true"}, idOpts[0])

	emailField, err := NewField(*u.Fields[6])
	assert.NoError(t, err)
	emailOpts := emailField.listFieldOptions()
	assert.Len(t, emailOpts, 2)
	assert.Equal(t, FieldOption{Name: "ddl.protobuf.v1.index", Value: "name=idx_email"}, emailOpts[1])
}

func TestToSQL(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	idField, err := NewField(*u.Fields[0])
	assert.NoError(t, err)
	assert.Equal(t, "id BIGINT UNSIGNED", idField.ToSQL())
}
