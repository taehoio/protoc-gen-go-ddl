package ddl

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
)

func gen() (*protogen.Plugin, error) {
	// https://medium.com/@tim.r.coulson/writing-a-protoc-plugin-with-google-golang-org-protobuf-cd5aa75f5777
	b, err := os.ReadFile("../../testdata/marshaled_input.dat")
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

	for _, file := range gen.Files {
		if file.GoPackageName == "testv1" {
			return file, nil
		}
	}

	return nil, errors.New("user file not found")
}

func userMessage() (*protogen.Message, error) {
	userFile, err := userFile()
	if err != nil {
		return nil, err
	}

	for _, msg := range userFile.Messages {
		if msg.GoIdent.GoName == "User" {
			return msg, nil
		}
	}

	return nil, errors.New("user message not found")
}

func userCheckinMessage() (*protogen.Message, error) {
	file, err := userFile()
	if err != nil {
		return nil, err
	}

	return file.Messages[1], nil
}

func TestKindToSQLType(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	type test struct {
		kind protoreflect.Kind
		want string
	}

	tests := []test{
		{protoreflect.Int64Kind, "BIGINT NOT NULL"},
		{protoreflect.MessageKind, "TIMESTAMP(6) NOT NULL"},
		{protoreflect.MessageKind, "TIMESTAMP(6) NOT NULL"},
		{protoreflect.MessageKind, "TIMESTAMP(6) NULL DEFAULT NULL"},
		{protoreflect.StringKind, "VARCHAR(255) NOT NULL"},
		{protoreflect.StringKind, "VARCHAR(100) NULL DEFAULT NULL"},
		{protoreflect.StringKind, "VARCHAR(255) NOT NULL"},
		{protoreflect.StringKind, "TEXT NOT NULL"},
		{protoreflect.MessageKind, "DATE NOT NULL"},
		{protoreflect.MessageKind, "DATE NULL DEFAULT NULL"},
		{protoreflect.StringKind, "VARCHAR(255) NOT NULL"},
	}

	for i, pbField := range u.Fields {
		ddlField, err := NewField(*pbField)
		assert.NoError(t, err)
		assert.NotEmpty(t, ddlField)

		tt := tests[i]
		assert.EqualValues(t, tt.kind, pbField.Desc.Kind())
		assert.EqualValues(t, tt.want, ddlField.SQLType)
	}
}

func TestListFieldOptions(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	idField := *u.Fields[0]
	assert.NoError(t, err)
	idOpts := listFieldOptions(idField)
	assert.Len(t, idOpts, 2)
	assert.Equal(t, FieldOption{Name: "taehoio.ddl.protobuf.v1.key", Value: "true"}, idOpts[0])
	assert.Equal(t, FieldOption{Name: "taehoio.ddl.protobuf.v1.type", Value: ""}, idOpts[1])

	emailField := *u.Fields[6]
	assert.NoError(t, err)
	emailOpts := listFieldOptions(emailField)
	assert.Len(t, emailOpts, 3)
	assert.Equal(t, FieldOption{Name: "taehoio.ddl.protobuf.v1.key", Value: "false"}, emailOpts[0])
	assert.Equal(t, FieldOption{Name: "taehoio.ddl.protobuf.v1.index", Value: "name=idx_email"}, emailOpts[1])
	assert.Equal(t, FieldOption{Name: "taehoio.ddl.protobuf.v1.type", Value: ""}, emailOpts[2])

	profileJSONField := *u.Fields[7]
	assert.NoError(t, err)
	profileJSONOpts := listFieldOptions(profileJSONField)
	assert.Len(t, profileJSONOpts, 2)
	assert.Equal(t, FieldOption{Name: "taehoio.ddl.protobuf.v1.type", Value: "name=TEXT"}, profileJSONOpts[1])
}

func TestToSQL(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	idField, err := NewField(*u.Fields[0])
	assert.NoError(t, err)
	assert.Equal(t, "`id` BIGINT NOT NULL", idField.ToDDLSQL())
}

func TestNewField(t *testing.T) {
	t.Run("with user_checkin message", func(t *testing.T) {
		userCheckin, err := userCheckinMessage()
		assert.NoError(t, err)

		t.Run("country_code field type is CountryCode", func(t *testing.T) {
			countryCodeField := userCheckin.Fields[4]
			assert.NotNil(t, countryCodeField)

			f, err := NewField(*countryCodeField)
			assert.NoError(t, err)

			assert.Equal(t, f.PbType, "CountryCode")
			assert.Equal(t, f.GoType, "CountryCode")
		})
	})
}
