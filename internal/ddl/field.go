package ddl

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	protobufv1 "github.com/taehoio/protoc-gen-go-ddl/gen/go/taehoio/ddl/protobuf/v1"
)

type Field struct {
	field protogen.Field

	Options []FieldOption

	TextName  string
	GoName    string
	VarName   string
	SQLName   string
	ParamName string

	PbKind        string
	PbType        string
	SQLType       string
	GoType        string
	GoValueType   string
	GoTypeVarName string

	SupportsSQLNullType bool
	IsWellKnownType     bool
	IsRepeatedType      bool
	IsEnumType          bool
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
	if field.Enum != nil {
		pbType = field.Enum.GoIdent.GoName
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

		TextName:  field.Desc.TextName(),
		GoName:    field.GoName,
		VarName:   strcase.ToLowerCamel(field.GoName),
		SQLName:   string(field.Desc.Name()),
		ParamName: strcase.ToLowerCamel(field.GoName + "Param"),

		PbKind:        field.Desc.Kind().String(),
		PbType:        pbType,
		SQLType:       sqlType,
		GoType:        goType,
		GoValueType:   goType,
		GoTypeVarName: strcase.ToLowerCamel(strings.TrimPrefix(goType, "*")),

		SupportsSQLNullType: false,
		IsWellKnownType:     isWellKnownFieldType(field.Desc),
		IsRepeatedType:      field.Desc.IsList(),
		IsEnumType:          field.Desc.Kind() == protoreflect.EnumKind,
	}

	f.GoValueType = strings.TrimPrefix(f.GoValueType, "*")

	if strings.HasPrefix(goType, "*") || strings.Contains(goType, ".") {
		f.SupportsSQLNullType = true
	}

	return f, nil
}

func (f Field) ToDDLSQL() string {
	return fmt.Sprintf("`%s` %s", f.SQLName, f.SQLType)
}

func listFieldOptions(f protogen.Field) []FieldOption {
	var fieldOpts []FieldOption

	opts := f.Desc.Options().(*descriptorpb.FieldOptions)

	if keyOptVal, ok := proto.GetExtension(opts, protobufv1.E_Key).(bool); ok {
		fieldOpts = append(fieldOpts, FieldOption{
			Name:  string(protobufv1.E_Key.TypeDescriptor().FullName()),
			Value: fmt.Sprintf("%v", keyOptVal),
		})
	}

	if indexOptVal, ok := proto.GetExtension(opts, protobufv1.E_Index).([]string); ok {
		for _, v := range indexOptVal {
			fieldOpts = append(fieldOpts, FieldOption{
				Name:  string(protobufv1.E_Index.TypeDescriptor().FullName()),
				Value: v,
			})
		}
	}

	if uniqueOptVal, ok := proto.GetExtension(opts, protobufv1.E_Unique).([]string); ok {
		for _, v := range uniqueOptVal {
			fieldOpts = append(fieldOpts, FieldOption{
				Name:  string(protobufv1.E_Unique.TypeDescriptor().FullName()),
				Value: v,
			})
		}
	}

	if typeOptVal, ok := proto.GetExtension(opts, protobufv1.E_Type).(string); ok {
		fieldOpts = append(fieldOpts, FieldOption{
			Name:  string(protobufv1.E_Type.TypeDescriptor().FullName()),
			Value: typeOptVal,
		})
	}

	return fieldOpts
}

func isOptionalField(f protogen.Field) bool {
	return f.Desc.HasOptionalKeyword()
}

func extractTypeNameFromFieldOption(f protogen.Field) (string, bool) {
	opts := f.Desc.Options().(*descriptorpb.FieldOptions)

	if typeOptVal, ok := proto.GetExtension(opts, protobufv1.E_Type).(string); ok && typeOptVal != "" {
		typeKVs := strings.Split(typeOptVal, ",")
		for _, typeKV := range typeKVs {
			kv := strings.Split(typeKV, "=")
			if len(kv) != 2 {
				continue
			}
			k := kv[0]
			v := kv[1]
			if k == "name" {
				if isOptionalField(f) {
					return fmt.Sprintf("%s NULL DEFAULT NULL", v), true
				}
				return fmt.Sprintf("%s NOT NULL", v), true
			}
		}
	}

	return "", false
}

func kindToSQLType(f protogen.Field) (string, error) {
	if typeName, ok := extractTypeNameFromFieldOption(f); ok {
		return typeName, nil
	}

	kind := f.Desc.Kind()

	switch kind {
	case protoreflect.Int32Kind:
		if isOptionalField(f) {
			return "INT NULL DEFAULT NULL", nil
		} else {
			return "INT NOT NULL", nil
		}
	case protoreflect.Int64Kind:
		if isOptionalField(f) {
			return "BIGINT NULL DEFAULT NULL", nil
		} else {
			return "BIGINT NOT NULL", nil
		}
	case protoreflect.Uint32Kind:
		if isOptionalField(f) {
			return "INT UNSIGNED NULL DEFAULT NULL", nil
		} else {
			return "INT UNSIGNED NOT NULL", nil
		}
	case protoreflect.Uint64Kind:
		if isOptionalField(f) {
			return "BIGINT UNSIGNED NULL DEFAULT NULL", nil
		} else {
			return "BIGINT UNSIGNED NOT NULL", nil
		}
	case protoreflect.FloatKind:
		if isOptionalField(f) {
			return "FLOAT NULL DEFAULT NULL", nil
		} else {
			return "FLOAT NOT NULL", nil
		}
	case protoreflect.DoubleKind:
		if isOptionalField(f) {
			return "DOUBLE NULL DEFAULT NULL", nil
		} else {
			return "DOUBLE NOT NULL", nil
		}
	case protoreflect.StringKind:
		if isOptionalField(f) {
			return fmt.Sprintf("VARCHAR(%d) NULL DEFAULT NULL", defaultVarcharLength), nil
		} else {
			return fmt.Sprintf("VARCHAR(%d) NOT NULL", defaultVarcharLength), nil
		}
	case protoreflect.BoolKind:
		if isOptionalField(f) {
			return "BOOLEAN NULL DEFAULT NULL", nil
		} else {
			return "BOOLEAN NOT NULL", nil
		}
	case protoreflect.EnumKind:
		if isOptionalField(f) {
			return "INT NULL DEFAULT NULL", nil
		} else {
			return "INT NOT NULL", nil
		}
	case protoreflect.BytesKind:
		if isOptionalField(f) {
			return "BLOB NULL DEFAULT NULL", nil
		} else {
			return "BLOB NOT NULL", nil
		}

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
		if isOptionalField(f) {
			return "TIMESTAMP(6) NULL DEFAULT NULL", nil
		}
		return "TIMESTAMP(6) NOT NULL", nil
	case "google.type.Date":
		if isOptionalField(f) {
			return "DATE NULL DEFAULT NULL", nil
		}
		return "DATE NOT NULL", nil
	default:
		return "JSONB", nil
	}
}

func kindToGoType(f protogen.Field) (string, error) {
	kind := f.Desc.Kind()

	switch kind {
	case protoreflect.Int32Kind:
		if isOptionalField(f) {
			return "*int32", nil
		}
		return "int32", nil
	case protoreflect.Int64Kind:
		if isOptionalField(f) {
			return "*int64", nil
		} else {
			return "int64", nil
		}
	case protoreflect.Uint32Kind:
		if isOptionalField(f) {
			return "*uint32", nil
		} else {
			return "uint32", nil
		}
	case protoreflect.Uint64Kind:
		if isOptionalField(f) {
			return "*uint64", nil
		} else {
			return "uint64", nil
		}
	case protoreflect.FloatKind:
		if isOptionalField(f) {
			return "*float32", nil
		} else {
			return "float32", nil
		}
	case protoreflect.DoubleKind:
		if isOptionalField(f) {
			return "*float64", nil
		} else {
			return "float64", nil
		}
	case protoreflect.StringKind:
		if isOptionalField(f) {
			return "*string", nil
		} else {
			return "string", nil
		}
	case protoreflect.BoolKind:
		if isOptionalField(f) {
			return "*bool", nil
		} else {
			return "bool", nil
		}
	case protoreflect.BytesKind:
		if isOptionalField(f) {
			return "*[]byte", nil
		} else {
			return "[]byte", nil
		}

	case protoreflect.EnumKind:
		if isOptionalField(f) {
			return "*" + f.Enum.GoIdent.GoName, nil
		} else {
			return f.Enum.GoIdent.GoName, nil
		}
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
		if isOptionalField(f) {
			return "sql.NullTime", nil
		}
		return "time.Time", nil
	case "google.type.Date":
		return "*date.Date", nil
	default:
		if isOptionalField(f) {
			return "*" + f.Message.GoIdent.GoName, nil
		}
		return f.Message.GoIdent.GoName, nil
	}
}

func isWellKnownFieldType(field protoreflect.FieldDescriptor) bool {
	if field.Kind() != protoreflect.MessageKind {
		return true
	}

	return field.Message().FullName().Parent() == "google.protobuf"
}
