package ddl

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	protobufv1 "github.com/taehoio/ddl/gen/go/ddl/protobuf/v1"
)

type Field struct {
	field protogen.Field

	Name    string
	Type    string
	Options []FieldOption
}

type FieldOption struct {
	Name  string
	Value string
}

func NewField(field protogen.Field) (*Field, error) {
	f := &Field{
		field: field,
	}

	f.Name = field.Desc.TextName()

	sqlType, err := f.kindToSQLType()
	if err != nil {
		return nil, err
	}
	f.Type = sqlType

	fieldOptions := f.listFieldOptions()
	f.Options = fieldOptions

	return f, nil
}

const defaultVarcharLength = 255

func (f Field) kindToSQLType() (string, error) {
	kind := f.field.Desc.Kind()

	switch kind {
	case protoreflect.Int32Kind:
		return "INT", nil
	case protoreflect.Int64Kind:
		return "BIGINT", nil
	case protoreflect.Uint32Kind:
		return "INT UNSIGNED", nil
	case protoreflect.Uint64Kind:
		return "BIGINT UNSIGNED", nil
	case protoreflect.FloatKind:
		return "FLOAT", nil
	case protoreflect.DoubleKind:
		return "DOUBLE", nil
	case protoreflect.StringKind:
		return fmt.Sprintf("VARCHAR(%d)", defaultVarcharLength), nil
	case protoreflect.BoolKind:
		return "BOOLEAN", nil
	case protoreflect.EnumKind:
		return "INT", nil
	case protoreflect.MessageKind:
		if f.field.Desc.Message().FullName() == "google.protobuf.Timestamp" {
			return "TIMESTAMP NULL DEFAULT NULL", nil
		}
		return "JSONB", nil
	default:
		return "", fmt.Errorf("unsupported kind %s", kind)
	}
}

func (f Field) listFieldOptions() []FieldOption {
	var fieldOptions []FieldOption

	opts := f.field.Desc.Options().(*descriptorpb.FieldOptions)

	keyOptVal := proto.GetExtension(opts, protobufv1.E_Key).(bool)

	fieldOptions = append(fieldOptions, FieldOption{
		Name:  string(protobufv1.E_Key.TypeDescriptor().FullName()),
		Value: fmt.Sprintf("%v", keyOptVal),
	})

	indexOptVal := proto.GetExtension(opts, protobufv1.E_Index).([]string)
	for _, v := range indexOptVal {
		fieldOptions = append(fieldOptions, FieldOption{
			Name:  string(protobufv1.E_Index.TypeDescriptor().FullName()),
			Value: v,
		})
	}

	return fieldOptions
}

func (f Field) ToSQL() string {
	return fmt.Sprintf("%s %s", f.Name, f.Type)
}
