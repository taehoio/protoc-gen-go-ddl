package ddl

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	protobufv1 "github.com/taehoio/ddl/gen/go/ddl/protobuf/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type MessageInfo struct {
	message        protogen.Message
	extensionTypes protoregistry.Types

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

func NewMessageInfo(message protogen.Message, extensionTypes protoregistry.Types) (*MessageInfo, error) {
	t := &MessageInfo{
		message:        message,
		extensionTypes: extensionTypes,
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

func (mi *MessageInfo) listMessageOptions() ([]MessageOption, error) {
	// The MessageOptions as provided by protoc does not know about
	// dynamically created extensions, so they are left as unknown fields.
	// We round-trip marshal and unmarshal the opts with
	// a dynamically created resolver that does know about extensions at runtime.
	opts := mi.message.Desc.Options().(*descriptorpb.MessageOptions)
	b, err := proto.Marshal(opts)
	if err != nil {
		return nil, err
	}
	opts.Reset()
	err = proto.UnmarshalOptions{Resolver: &mi.extensionTypes}.Unmarshal(b, opts)
	if err != nil {
		return nil, err
	}

	var messageOptions []MessageOption

	// Use protobuf reflection to iterate over all the extension fields,
	// looking for the ones that we are interested in.
	opts.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsExtension() {
			return true
		}

		messageOptions = append(messageOptions, MessageOption{
			Name:  string(fd.Name()),
			Value: v.String(),
		})

		return true
	})

	return messageOptions, nil
}

func (mi *MessageInfo) extractFields() ([]Field, error) {
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

func (mi *MessageInfo) extractKeys() ([]string, error) {
	var keys []string

	for _, field := range mi.Fields {
		for _, opt := range field.Options {
			if opt.Name == protobufv1.E_Key.Name && opt.Value == "true" {
				keys = append(keys, field.Name)
			}
		}
	}

	return keys, nil
}

func (mi *MessageInfo) extractIndices() ([]Index, error) {
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

func (mi *MessageInfo) GenerateDDLToCreate() (string, error) {
	tableName := strcase.ToSnake(string(mi.message.Desc.Name()))

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
