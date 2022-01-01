package ddl

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	protobufv1 "github.com/taehoio/ddl/gen/go/ddl/protobuf/v1"
)

type MessageInfo struct {
	message protogen.Message

	MessageOptions []MessageOption
	Fields         []Field

	Keys    []string
	Indices []Index
}

type MessageOption struct {
	Name  string
	Value string
}

type Index struct {
	Name       string
	FieldNames []string
}

func NewMessageInfo(message protogen.Message) (*MessageInfo, error) {
	t := &MessageInfo{
		message: message,
	}

	messageOptions, err := t.listMessageOptions()
	if err != nil {
		return nil, err
	}
	t.MessageOptions = messageOptions

	fields, err := t.extractFields()
	if err != nil {
		return nil, err
	}
	t.Fields = fields

	keys, err := t.extractKeys()
	if err != nil {
		return nil, err
	}
	t.Keys = keys

	indices, err := t.extractIndices()
	if err != nil {
		return nil, err
	}
	t.Indices = indices

	return t, nil
}

func (mi MessageInfo) listMessageOptions() ([]MessageOption, error) {
	var messageOptions []MessageOption

	opts := mi.message.Desc.Options().(*descriptorpb.MessageOptions)

	datastoreTypeOptVal := proto.GetExtension(opts, protobufv1.E_DatastoreType).(protobufv1.DatastoreType)

	messageOptions = append(messageOptions, MessageOption{
		Name:  string(protobufv1.E_DatastoreType.TypeDescriptor().FullName()),
		Value: datastoreTypeOptVal.String(),
	})

	return messageOptions, nil
}

func (mi MessageInfo) extractFields() ([]Field, error) {
	var fields []Field

	for _, field := range mi.message.Fields {
		f, err := NewField(*field)
		if err != nil {
			return nil, err
		}

		fields = append(fields, *f)
	}

	return fields, nil
}

func (mi MessageInfo) extractKeys() ([]string, error) {
	var keys []string

	for _, field := range mi.Fields {
		for _, opt := range field.Options {
			if opt.Name == string(protobufv1.E_Key.TypeDescriptor().FullName()) && opt.Value == "true" {
				keys = append(keys, field.Name)
			}
		}
	}

	return keys, nil
}

func (mi MessageInfo) extractIndices() ([]Index, error) {
	indexMap := make(map[string][]string)

	for _, field := range mi.Fields {
		for _, opt := range field.Options {
			if opt.Name == protobufv1.E_Index.Name {
				kvPairs := strings.Split(opt.Value, ",")
				for _, kvPair := range kvPairs {
					kv := strings.Split(kvPair, "=")
					k := kv[0]
					indexName := kv[1]
					if k == "name" {
						indexMap[indexName] = append(indexMap[indexName], field.Name)
					}
				}
			}
		}
	}

	var indices []Index
	for k, v := range indexMap {
		indices = append(indices, Index{
			Name:       k,
			FieldNames: v,
		})
	}

	return indices, nil
}

var (
	ErrNotSupportedDatastore = fmt.Errorf("not supported datastore")
)

func (mi MessageInfo) GenerateDDLSQL() (string, error) {
	tableName := strcase.ToSnake(string(mi.message.Desc.Name()))

	if !mi.supportsMySQL() {
		return "", ErrNotSupportedDatastore
	}

	var ddlFields []string
	for _, field := range mi.Fields {
		ddlFields = append(ddlFields, field.ToSQL())
	}

	ddlCreateTable := fmt.Sprintf("CREATE TABLE %s (\n\t%v,\n\tPRIMARY KEY (%s)\n);", tableName, strings.Join(ddlFields, ",\n\t"), strings.Join(mi.Keys, ", "))

	var stmts []string
	stmts = append(stmts, ddlCreateTable)

	for _, index := range mi.Indices {
		stmts = append(stmts, fmt.Sprintf("CREATE INDEX %s ON %s (%s);", index.Name, tableName, strings.Join(index.FieldNames, ", ")))
	}

	return strings.Join(stmts, "\n\n"), nil
}

func (mi MessageInfo) supportsMySQL() bool {
	for _, opt := range mi.MessageOptions {
		if opt.Name == string(protobufv1.E_DatastoreType.TypeDescriptor().FullName()) && opt.Value == protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL.String() {
			return true
		}
	}
	return false
}
