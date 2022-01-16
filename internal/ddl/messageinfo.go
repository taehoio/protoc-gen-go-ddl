package ddl

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	protobufv1 "github.com/taehoio/ddl/gen/go/taehoio/ddl/protobuf/v1"
)

var (
	//go:embed template/ddl.pb.go.tmpl
	dmlTmpl string
)

type MessageInfo struct {
	message protogen.Message

	MessageOptions []MessageOption
	Fields         []Field

	Keys    []string
	Indices []Index
	Uniques []Unique
}

type MessageOption struct {
	Name  string
	Value string
}

type Index struct {
	Name       string
	FieldNames []string
}

type Unique struct {
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

	uniques, err := t.extractUniques()
	if err != nil {
		return nil, err
	}
	t.Uniques = uniques

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
			if opt.Name == string(protobufv1.E_Index.TypeDescriptor().FullName()) {
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

func (mi MessageInfo) extractUniques() ([]Unique, error) {
	uniqueMap := make(map[string][]string)

	for _, field := range mi.Fields {
		for _, opt := range field.Options {
			if opt.Name == string(protobufv1.E_Unique.TypeDescriptor().FullName()) {
				kvPairs := strings.Split(opt.Value, ",")
				for _, kvPair := range kvPairs {
					kv := strings.Split(kvPair, "=")
					k := kv[0]
					uniqueName := kv[1]
					if k == "name" {
						uniqueMap[uniqueName] = append(uniqueMap[uniqueName], field.Name)
					}
				}
			}
		}
	}

	var uniques []Unique
	for k, v := range uniqueMap {
		uniques = append(uniques, Unique{
			Name:       k,
			FieldNames: v,
		})
	}

	return uniques, nil
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

	var keys []string
	for _, k := range mi.Keys {
		keys = append(keys, fmt.Sprintf("`%s`", k))
	}

	ddlCreateTable := fmt.Sprintf("\nCREATE TABLE `%s` (\n\t%v,\n\tPRIMARY KEY (%s)\n);", tableName, strings.Join(ddlFields, ",\n\t"), strings.Join(keys, ", "))

	var stmts []string
	stmts = append(stmts, ddlCreateTable)

	for _, index := range mi.Indices {
		var indexFieldNames []string
		for _, ifn := range index.FieldNames {
			indexFieldNames = append(indexFieldNames, fmt.Sprintf("`%s`", ifn))
		}

		stmts = append(stmts, fmt.Sprintf("\nCREATE INDEX `%s` ON `%s` (%s);", index.Name, tableName, strings.Join(indexFieldNames, ", ")))
	}

	for _, unique := range mi.Uniques {
		var uniqueFieldNames []string
		for _, ufn := range unique.FieldNames {
			uniqueFieldNames = append(uniqueFieldNames, fmt.Sprintf("`%s`", ufn))
		}

		stmts = append(stmts, fmt.Sprintf("\nCREATE UNIQUE INDEX `%s` ON `%s` (%s);", unique.Name, tableName, strings.Join(uniqueFieldNames, ", ")))

	}

	return strings.Join(stmts, "\n"), nil
}

func (mi MessageInfo) supportsMySQL() bool {
	for _, opt := range mi.MessageOptions {
		if opt.Name == string(protobufv1.E_DatastoreType.TypeDescriptor().FullName()) && opt.Value == protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL.String() {
			return true
		}
	}
	return false
}

type dml struct {
	Pkg dmlPackage
	Msg dmlMessage
}

type dmlPackage struct {
	Name string
}

type dmlMessage struct {
	Name      string
	VarName   string
	TableName string
	Fields    []dmlField
	KeyFields []dmlField
	Indices   []Index
}

type dmlField struct {
	Name                  string
	VarName               string
	Kind                  string
	Type                  string
	ShouldWrapWithSQLType bool
	SQLType               string
	SQLName               string
}

func (mi MessageInfo) GenerateDMLSQL() (string, error) {
	tmpl, err := template.New("dmlFile").Parse(dmlTmpl)
	if err != nil {
		return "", err
	}

	pkgName := ""
	goPkg := mi.message.Desc.ParentFile().Options().(*descriptorpb.FileOptions).GetGoPackage()
	goPkgSplitted := strings.Split(goPkg, ";")
	if len(goPkgSplitted) >= 2 {
		pkgName = goPkgSplitted[len(goPkgSplitted)-1]
	} else {
		goPkgSplitted = strings.Split(goPkg, "/")
		pkgName = goPkgSplitted[len(goPkgSplitted)-1]
	}

	keys, err := mi.extractKeys()
	if err != nil {
		return "", err
	}

	indices, err := mi.extractIndices()
	if err != nil {
		return "", err
	}

	var dmlFields []dmlField
	var keyFields []dmlField

	for _, field := range mi.Fields {
		fieldType := ""
		if field.field.Desc.Message() != nil {
			fieldType = string(field.field.Desc.Message().FullName())
		}

		shouldWrapWithSQLType := false
		sqlType := ""
		if field.field.Desc.Kind() == protoreflect.MessageKind {
			shouldWrapWithSQLType = true

			if fieldType == "google.protobuf.Timestamp" {
				sqlType = "sql.NullTime"
			}
			if fieldType == "google.protobuf.StringValue" {
				sqlType = "sql.NullString"
			}
		}

		isKey := false
		for _, key := range keys {
			if key == field.Name {
				isKey = true
				break
			}
		}

		f := dmlField{
			Name:                  field.field.GoName,
			VarName:               strcase.ToLowerCamel(field.field.GoName),
			Kind:                  field.field.Desc.Kind().String(),
			Type:                  fieldType,
			ShouldWrapWithSQLType: shouldWrapWithSQLType,
			SQLType:               sqlType,
			SQLName:               strcase.ToSnake(field.field.GoName),
		}

		dmlFields = append(dmlFields, f)
		if isKey {
			keyFields = append(keyFields, f)
		}
	}

	for i, index := range indices {
		var camelCaseFieldNames []string
		for _, fieldName := range index.FieldNames {
			camelCaseFieldNames = append(camelCaseFieldNames, strcase.ToCamel(fieldName))
		}
		indices[i].Name = strings.Join(camelCaseFieldNames, "And")
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, &dml{
		Pkg: dmlPackage{
			Name: pkgName,
		},
		Msg: dmlMessage{
			Name:      strcase.ToCamel(string(mi.message.Desc.Name())),
			VarName:   strcase.ToLowerCamel(string(mi.message.Desc.Name())),
			TableName: strcase.ToSnake((string(mi.message.Desc.Name()))),
			Fields:    dmlFields,
			KeyFields: keyFields,
			Indices:   indices,
		},
	}); err != nil {
		return "", err
	}

	return b.String(), nil
}
