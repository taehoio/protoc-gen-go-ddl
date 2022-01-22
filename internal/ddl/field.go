package ddl

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/iancoleman/strcase"

	protobufv1 "github.com/taehoio/ddl/gen/go/taehoio/ddl/protobuf/v1"
)

type Field struct {
	field protogen.Field

	Options []FieldOption

	TextName string
	GoName   string
	VarName  string
	SQLName  string

	PbKind  string
	PbType  string
	SQLType string
	GoType  string

	SupportsSQLNullType bool
}

type FieldOption struct {
	Name  string
	Value string
}

func NewField(field protogen.Field) (*Field, error) {
	pbType := ""
	if field.Desc.Message() != nil {
		pbType = string(field.Desc.Message().FullName())
	}

	sqlType, err := kindToSQLType(field)
	if err != nil {
		return nil, err
	}

	goType, err := kindToGoType(field)
	if err != nil {
		return nil, err
	}

	f := &Field{
		field: field,

		Options: listFieldOptions(field),

		TextName: field.Desc.TextName(),
		GoName:   field.GoName,
		VarName:  strcase.ToLowerCamel(field.GoName),
		SQLName:  strcase.ToSnake(field.GoName),

		PbKind:  field.Desc.Kind().String(),
		PbType:  pbType,
		SQLType: sqlType,
		GoType:  goType,

		SupportsSQLNullType: false,
	}

	if strings.Contains(goType, ".") {
		f.SupportsSQLNullType = true
	}

	return f, nil
}

func (f Field) ToDDLSQL() string {
	return fmt.Sprintf("`%s` %s", f.TextName, f.SQLType)
}

func listFieldOptions(f protogen.Field) []FieldOption {
	var fieldOptions []FieldOption

	opts := f.Desc.Options().(*descriptorpb.FieldOptions)

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

	uniqueOptVal := proto.GetExtension(opts, protobufv1.E_Unique).([]string)
	for _, v := range uniqueOptVal {
		fieldOptions = append(fieldOptions, FieldOption{
			Name:  string(protobufv1.E_Unique.TypeDescriptor().FullName()),
			Value: v,
		})
	}

	return fieldOptions
}

func kindToSQLType(f protogen.Field) (string, error) {
	kind := f.Desc.Kind()

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
		return messageKindToSQLType(f)
	default:
		return "", fmt.Errorf("unsupported kind %s", kind)
	}
}

func messageKindToSQLType(f protogen.Field) (string, error) {
	if f.Desc.Kind() != protoreflect.MessageKind {
		return "", fmt.Errorf("not message kind but it is %s", f.Desc.Kind())
	}

	name := f.Desc.Message().FullName()
	switch name {
	case "google.protobuf.Timestamp":
		return "TIMESTAMP(6) NULL DEFAULT NULL", nil
	case "google.protobuf.StringValue":
		return fmt.Sprintf("VARCHAR(%d)", defaultVarcharLength), nil
	case "google.protobuf.Int32Value":
		return "INT", nil
	case "google.protobuf.Int64Value":
		return "BIGINT", nil
	case "google.protobuf.UInt32Value":
		return "INT UNSIGNED", nil
	case "google.protobuf.UInt64Value":
		return "BIGINT UNSIGNED", nil
	case "google.protobuf.FloatValue":
		return "FLOAT", nil
	case "google.protobuf.DoubleValue":
		return "DOUBLE", nil
	case "google.protobuf.BoolValue":
		return "BOOLEAN", nil
	case "google.protobuf.BytesValue":
		return "BLOB", nil
	default:
		return "JSONB", nil
	}
}

func kindToGoType(f protogen.Field) (string, error) {
	kind := f.Desc.Kind()

	switch kind {
	case protoreflect.Int32Kind:
		return "int32", nil
	case protoreflect.Int64Kind:
		return "int64", nil
	case protoreflect.Uint32Kind:
		return "uint32", nil
	case protoreflect.Uint64Kind:
		return "uint64", nil
	case protoreflect.FloatKind:
		return "float32", nil
	case protoreflect.DoubleKind:
		return "float64", nil
	case protoreflect.StringKind:
		return "string", nil
	case protoreflect.BoolKind:
		return "bool", nil
	case protoreflect.EnumKind:
		return "int", nil
	case protoreflect.MessageKind:
		return messageKindToGoType(f)
	default:
		return "", fmt.Errorf("unsupported kind %s", kind)
	}
}

func messageKindToGoType(f protogen.Field) (string, error) {
	if f.Desc.Kind() != protoreflect.MessageKind {
		return "", fmt.Errorf("not message kind but it is %s", f.Desc.Kind())
	}

	name := f.Desc.Message().FullName()
	switch name {
	case "google.protobuf.Timestamp":
		return "sql.NullTime", nil
	case "google.protobuf.StringValue":
		return "sql.NullString", nil
	default:
		return "", fmt.Errorf("not supported message %s", f.Desc.Message().FullName())
	}
}
