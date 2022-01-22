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
	"google.golang.org/protobuf/types/descriptorpb"

	protobufv1 "github.com/taehoio/ddl/gen/go/taehoio/ddl/protobuf/v1"
)

var (
	//go:embed template/dml_message.pb.go.tmpl
	dmlMessageTmpl string
)

type MessageInfo struct {
	message protogen.Message

	MessageOptions []MessageOption

	Name    string
	VarName string
	GoName  string
	SQLName string

	Fields    []Field
	KeyFields []Field
	Indices   []Index
	Uniques   []Unique
}

type MessageOption struct {
	Name  string
	Value string
}

type Index struct {
	Name   string
	Fields []Field
}

type Unique struct {
	Name   string
	Fields []Field
}

func NewMessageInfo(message protogen.Message) (*MessageInfo, error) {
	messageOptions, err := listMessageOptions(message)
	if err != nil {
		return nil, err
	}

	fields, err := extractFields(message)
	if err != nil {
		return nil, err
	}

	keyFields, err := extractKeyFields(fields)
	if err != nil {
		return nil, err
	}

	indices, err := extractIndices(fields)
	if err != nil {
		return nil, err
	}

	uniques, err := extractUniques(fields)
	if err != nil {
		return nil, err
	}

	messageName := string(message.Desc.Name())

	mi := &MessageInfo{
		message: message,

		MessageOptions: messageOptions,

		Name:    messageName,
		VarName: strcase.ToLowerCamel(messageName),
		GoName:  strcase.ToCamel(messageName),
		SQLName: strcase.ToSnake(messageName),

		Fields:    fields,
		KeyFields: keyFields,
		Indices:   indices,
		Uniques:   uniques,
	}

	return mi, nil
}

func listMessageOptions(m protogen.Message) ([]MessageOption, error) {
	var messageOptions []MessageOption

	opts := m.Desc.Options().(*descriptorpb.MessageOptions)

	datastoreTypeOptVal := proto.GetExtension(opts, protobufv1.E_DatastoreType).(protobufv1.DatastoreType)

	messageOptions = append(messageOptions, MessageOption{
		Name:  string(protobufv1.E_DatastoreType.TypeDescriptor().FullName()),
		Value: datastoreTypeOptVal.String(),
	})

	return messageOptions, nil
}

func extractFields(m protogen.Message) ([]Field, error) {
	var fields []Field

	for _, field := range m.Fields {
		f, err := NewField(*field)
		if err != nil {
			return nil, err
		}

		fields = append(fields, *f)
	}

	return fields, nil
}

func extractKeyFields(fields []Field) ([]Field, error) {
	var keyFields []Field

	for _, field := range fields {
		for _, opt := range field.Options {
			if opt.Name == string(protobufv1.E_Key.TypeDescriptor().FullName()) && opt.Value == "true" {
				keyFields = append(keyFields, field)
			}
		}
	}

	return keyFields, nil
}

func extractIndices(fields []Field) ([]Index, error) {
	indexMap := make(map[string][]Field)

	for _, field := range fields {
		for _, opt := range field.Options {
			if opt.Name == string(protobufv1.E_Index.TypeDescriptor().FullName()) {
				kvPairs := strings.Split(opt.Value, ",")
				for _, kvPair := range kvPairs {
					kv := strings.Split(kvPair, "=")
					k := kv[0]
					indexName := kv[1]
					if k == "name" {
						indexMap[indexName] = append(indexMap[indexName], field)
					}
				}
			}
		}
	}

	var indices []Index
	for k, v := range indexMap {
		indices = append(indices, Index{
			Name:   k,
			Fields: v,
		})
	}

	return indices, nil
}

func extractUniques(fields []Field) ([]Unique, error) {
	uniqueMap := make(map[string][]Field)

	for _, field := range fields {
		for _, opt := range field.Options {
			if opt.Name == string(protobufv1.E_Unique.TypeDescriptor().FullName()) {
				kvPairs := strings.Split(opt.Value, ",")
				for _, kvPair := range kvPairs {
					kv := strings.Split(kvPair, "=")
					k := kv[0]
					uniqueName := kv[1]
					if k == "name" {
						uniqueMap[uniqueName] = append(uniqueMap[uniqueName], field)
					}
				}
			}
		}
	}

	var uniques []Unique
	for k, v := range uniqueMap {
		uniques = append(uniques, Unique{
			Name:   k,
			Fields: v,
		})
	}

	return uniques, nil
}

var (
	ErrNotSupportedDatastore = fmt.Errorf("not supported datastore")
)

func (mi MessageInfo) GenerateDDLSQL() (string, error) {
	tableName := mi.SQLName

	if !mi.supportsMySQL() {
		return "", ErrNotSupportedDatastore
	}

	var ddlFields []string
	for _, field := range mi.Fields {
		ddlFields = append(ddlFields, field.ToDDLSQL())
	}

	var keys []string
	for _, k := range mi.KeyFields {
		keys = append(keys, fmt.Sprintf("`%s`", k.SQLName))
	}

	ddlCreateTable := fmt.Sprintf("\nCREATE TABLE `%s` (\n\t%v,\n\tPRIMARY KEY (%s)\n);", tableName, strings.Join(ddlFields, ",\n\t"), strings.Join(keys, ", "))

	var stmts []string
	stmts = append(stmts, ddlCreateTable)

	for _, index := range mi.Indices {
		var indexFieldNames []string
		for _, ifn := range index.Fields {
			indexFieldNames = append(indexFieldNames, fmt.Sprintf("`%s`", ifn.SQLName))
		}

		stmts = append(stmts, fmt.Sprintf("\nCREATE INDEX `%s` ON `%s` (%s);", index.Name, tableName, strings.Join(indexFieldNames, ", ")))
	}

	for _, unique := range mi.Uniques {
		var uniqueFieldNames []string
		for _, ufn := range unique.Fields {
			uniqueFieldNames = append(uniqueFieldNames, fmt.Sprintf("`%s`", ufn.SQLName))
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
	PackageName string
	Message     MessageInfo
}

func getGoPackageName(opts *descriptorpb.FileOptions) string {
	pkgName := ""

	goPkg := opts.GetGoPackage()
	goPkgSplitted := strings.Split(goPkg, ";")
	if len(goPkgSplitted) >= 2 {
		pkgName = goPkgSplitted[len(goPkgSplitted)-1]
	} else {
		goPkgSplitted = strings.Split(goPkg, "/")
		pkgName = goPkgSplitted[len(goPkgSplitted)-1]
	}

	return pkgName
}

func (mi MessageInfo) GenerateDMLSQL() (string, error) {
	tmpl, err := template.New("dmlMessageTmpl").Parse(dmlMessageTmpl)
	if err != nil {
		return "", err
	}

	pkgName := getGoPackageName(mi.message.Desc.ParentFile().Options().(*descriptorpb.FileOptions))

	var b bytes.Buffer
	if err := tmpl.Execute(&b, &dml{
		PackageName: pkgName,
		Message:     mi,
	}); err != nil {
		return "", err
	}

	return b.String(), nil
}
