package ddl

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	protobufv1 "github.com/taehoio/protoc-gen-go-ddl/gen/go/taehoio/ddl/protobuf/v1"
)

var (
	//go:embed template/dml_message_mysql.pb.go.tmpl
	dmlMessageMySQLTmpl string

	//go:embed template/dml_message_mongodb.pb.go.tmpl
	dmlMessageMongoDBTmpl string
)

var (
	dmlMessageTmpl = map[protobufv1.DatastoreType]string{
		protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL:   dmlMessageMySQLTmpl,
		protobufv1.DatastoreType_DATASTORE_TYPE_MONGODB: dmlMessageMongoDBTmpl,
	}
)

type MessageInfo struct {
	message protogen.Message

	MessageOptions []MessageOption
	NestedMessages []*MessageInfo

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

	nestedMessages, err := listNestedMessages(message)
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

	messageName := message.GoIdent.GoName

	mi := &MessageInfo{
		message: message,

		MessageOptions: messageOptions,
		NestedMessages: nestedMessages,

		Name:    messageName,
		VarName: strcase.ToLowerCamel(messageName),
		GoName:  strcase.ToCamel(messageName),
		SQLName: strcase.ToSnake(messageName),

		Fields:    fields,
		KeyFields: keyFields,
		Indices:   indices,
		Uniques:   uniques,
	}

	if datastore := mi.getDatastoreOption(); datastore != protobufv1.DatastoreType_DATASTORE_TYPE_UNSPECIFIED && datastore != protobufv1.DatastoreType_DATASTORE_TYPE_MONGODB && len(mi.NestedMessages) > 0 {
		return nil, fmt.Errorf("nested message in %s datastore is not supported", datastore.String())
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

func listNestedMessages(m protogen.Message) ([]*MessageInfo, error) {
	var nestedMessages []*MessageInfo

	for _, msg := range m.Messages {
		nestedMessage, err := NewMessageInfo(*msg)
		if err != nil {
			return nil, err
		}

		nestedMessages = append(nestedMessages, nestedMessage)
	}

	return nestedMessages, nil
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

type indexInfo struct {
	name  string
	field Field
	order int32
}

func extractIndices(fields []Field) ([]Index, error) {
	indexMap := make(map[string][]indexInfo)

	for fieldIdx, field := range fields {
		for _, opt := range field.Options {
			if opt.Name == string(protobufv1.E_Index.TypeDescriptor().FullName()) {
				ii, err := extractIndexInfo(field, int32(fieldIdx), opt.Value)
				if err != nil {
					return nil, err
				}

				indexMap[ii.name] = append(indexMap[ii.name], *ii)
			}
		}
	}

	var indices []Index
	for indexName, indexInfos := range indexMap {
		sort.Slice(indexInfos, func(i, j int) bool {
			return indexInfos[i].order < indexInfos[j].order
		})

		var fields []Field
		for _, ii := range indexInfos {
			fields = append(fields, ii.field)
		}

		indices = append(indices, Index{
			Name:   indexName,
			Fields: fields,
		})
	}

	sort.Slice(indices, func(i, j int) bool {
		return strings.Compare(indices[i].Name, indices[j].Name) < 0
	})

	return indices, nil
}

func extractUniques(fields []Field) ([]Unique, error) {
	uniqueMap := make(map[string][]indexInfo)

	for fieldIdx, field := range fields {
		for _, opt := range field.Options {
			if opt.Name == string(protobufv1.E_Unique.TypeDescriptor().FullName()) {
				ii, err := extractIndexInfo(field, int32(fieldIdx), opt.Value)
				if err != nil {
					return nil, err
				}

				uniqueMap[ii.name] = append(uniqueMap[ii.name], *ii)
			}
		}
	}

	var uniques []Unique
	for indexName, indexInfos := range uniqueMap {
		sort.Slice(indexInfos, func(i, j int) bool {
			return indexInfos[i].order < indexInfos[j].order
		})

		var fields []Field
		for _, ii := range indexInfos {
			fields = append(fields, ii.field)
		}

		uniques = append(uniques, Unique{
			Name:   indexName,
			Fields: fields,
		})
	}

	sort.Slice(uniques, func(i, j int) bool {
		return strings.Compare(uniques[i].Name, uniques[j].Name) < 0
	})

	return uniques, nil
}

var (
	ErrorDatastoreUnspecified = fmt.Errorf("datastore unspecified")
	ErrNotSupportedDatastore  = fmt.Errorf("not supported datastore")
)

func (mi MessageInfo) GenerateDDL() (string, error) {
	switch mi.getDatastoreOption() {
	case protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL:
		return mi.generateMySQLDDL()
	case protobufv1.DatastoreType_DATASTORE_TYPE_MONGODB:
		return mi.generateMongodbDDL()
	case protobufv1.DatastoreType_DATASTORE_TYPE_UNSPECIFIED:
		return "", ErrorDatastoreUnspecified
	default:
		return "", ErrNotSupportedDatastore
	}
}

// generateMySQLDDL generates ddl for mysql. MessageInfo must support mysql datastore.
func (mi MessageInfo) generateMySQLDDL() (string, error) {
	tableName := mi.SQLName

	var ddlFields []string
	for _, field := range mi.Fields {
		ddlFields = append(ddlFields, field.ToDDLSQL())
	}

	var keys []string
	for _, k := range mi.KeyFields {
		keys = append(keys, fmt.Sprintf("`%s`", k.SQLName))
	}

	var stmts []string

	ddlDropTable := fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName)
	stmts = append(stmts, ddlDropTable)

	ddlCreateTable := fmt.Sprintf("CREATE TABLE `%s` (\n\t%v,\n\tPRIMARY KEY (%s)\n);", tableName, strings.Join(ddlFields, ",\n\t"), strings.Join(keys, ", "))
	stmts = append(stmts, ddlCreateTable)

	for _, index := range mi.Indices {
		var indexFieldNames []string
		for _, ifn := range index.Fields {
			indexFieldNames = append(indexFieldNames, fmt.Sprintf("`%s`", ifn.SQLName))
		}

		stmts = append(stmts, fmt.Sprintf("CREATE INDEX `%s` ON `%s` (%s);", index.Name, tableName, strings.Join(indexFieldNames, ", ")))
	}

	for _, unique := range mi.Uniques {
		var uniqueFieldNames []string
		for _, ufn := range unique.Fields {
			uniqueFieldNames = append(uniqueFieldNames, fmt.Sprintf("`%s`", ufn.SQLName))
		}

		stmts = append(stmts, fmt.Sprintf("CREATE UNIQUE INDEX `%s` ON `%s` (%s);", unique.Name, tableName, strings.Join(uniqueFieldNames, ", ")))
	}

	return strings.Join(stmts, "\n\n"), nil
}

// generateMongodbDDL generates ddl for mongodb. MessageInfo must support mongodb datastore.
func (mi MessageInfo) generateMongodbDDL() (string, error) {
	collectionName := mi.SQLName

	var stmts []string
	stmts = append(stmts, fmt.Sprintf(`db.createCollection("%s")`, collectionName))

	var keys []string
	for _, k := range mi.KeyFields {
		keys = append(keys, fmt.Sprintf(`"%s":1`, k.SQLName))
	}
	stmts = append(stmts, fmt.Sprintf(`db.%s.createIndex({%s})`, collectionName, strings.Join(keys, ",")))

	for _, index := range mi.Indices {
		var indexFieldNames []string
		for _, field := range index.Fields {
			indexFieldNames = append(indexFieldNames, fmt.Sprintf(`"%s":1`, field.SQLName))
		}

		stmts = append(stmts, fmt.Sprintf(`db.%s.createIndex({%s}, {"name":"%s"})`, collectionName, strings.Join(indexFieldNames, ","), index.Name))
	}

	for _, unique := range mi.Uniques {
		var uniqFieldNames []string
		for _, field := range unique.Fields {
			uniqFieldNames = append(uniqFieldNames, fmt.Sprintf(`"%s":1`, field.SQLName))
		}

		stmts = append(stmts, fmt.Sprintf(`db.%s.createIndex({%s}, {"name":"%s","unique":true})`, collectionName, strings.Join(uniqFieldNames, ","), unique.Name))
	}

	return strings.Join(stmts, "\n"), nil
}

func (mi MessageInfo) supportsDatastore(datastore protobufv1.DatastoreType) bool {
	for _, opt := range mi.MessageOptions {
		if opt.Name == string(protobufv1.E_DatastoreType.TypeDescriptor().FullName()) && opt.Value == datastore.String() {
			return true
		}
	}
	return false
}

func (mi MessageInfo) getDatastoreOption() protobufv1.DatastoreType {
	for _, opt := range mi.MessageOptions {
		if opt.Name == string(protobufv1.E_DatastoreType.TypeDescriptor().FullName()) {
			return protobufv1.DatastoreType(protobufv1.DatastoreType_value[opt.Value])
		}
	}
	return protobufv1.DatastoreType_DATASTORE_TYPE_UNSPECIFIED
}

// DMLFileSuffix returns file suffix for dml file.
func (mi MessageInfo) DMLFileSuffix() (string, error) {
	switch mi.getDatastoreOption() {
	case protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL:
		return fmt.Sprintf("_dml_%s_mysql.pb.go", strcase.ToSnake(mi.Name)), nil
	case protobufv1.DatastoreType_DATASTORE_TYPE_MONGODB:
		return fmt.Sprintf("_dml_%s_mongodb.pb.go", strcase.ToSnake(mi.Name)), nil
	default:
		return "", ErrNotSupportedDatastore
	}
}

// DMLMockFileSuffix returns file suffix for dml mock file.
func (mi MessageInfo) DMLMockFileSuffix() (string, error) {
	switch mi.getDatastoreOption() {
	case protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL:
		return fmt.Sprintf("_dml_%s_mysql_mock.pb.go", strcase.ToSnake(mi.Name)), nil
	case protobufv1.DatastoreType_DATASTORE_TYPE_MONGODB:
		return fmt.Sprintf("_dml_%s_mongodb_mock.pb.go", strcase.ToSnake(mi.Name)), nil
	case protobufv1.DatastoreType_DATASTORE_TYPE_UNSPECIFIED:
		return "", ErrorDatastoreUnspecified
	default:
		return "", ErrNotSupportedDatastore
	}
}

// DDLFileSuffix returns the suffix of ddl file.
func (mi MessageInfo) DDLFileSuffix() (string, error) {
	switch mi.getDatastoreOption() {
	case protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL:
		return fmt.Sprintf("_ddl_%s_mysql.pb.sql", strcase.ToSnake(mi.Name)), nil
	case protobufv1.DatastoreType_DATASTORE_TYPE_MONGODB:
		return fmt.Sprintf("_ddl_%s_mongodb.pb.js", strcase.ToSnake(mi.Name)), nil
	default:
		return "", ErrNotSupportedDatastore
	}
}

type dml struct {
	PackageName  string
	GoImportPath string
	GoFilename   string
	MockFilename string
	Message      MessageInfo
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

func (mi MessageInfo) GenerateDML(goFilename string, mockFilename string) (string, error) {
	datastore := mi.getDatastoreOption()

	if _, ok := dmlMessageTmpl[datastore]; !ok {
		return "", ErrNotSupportedDatastore
	}

	tmpl, err := template.New("dmlMessageTmpl").Parse(dmlMessageTmpl[datastore])
	if err != nil {
		return "", err
	}

	pkgName := getGoPackageName(mi.message.Desc.ParentFile().Options().(*descriptorpb.FileOptions))

	var b bytes.Buffer
	if err := tmpl.Execute(&b, &dml{
		PackageName:  pkgName,
		GoImportPath: mi.message.GoIdent.GoImportPath.String(),
		GoFilename:   goFilename,
		MockFilename: mockFilename,
		Message:      mi,
	}); err != nil {
		return "", err
	}

	return b.String(), nil
}

func extractIndexInfo(field Field, fieldIdx int32, indexSpec string) (*indexInfo, error) {
	ii := &indexInfo{
		field: field,
		order: fieldIdx,
	}

	kvPairs := strings.Split(indexSpec, ",")
	for _, kvPair := range kvPairs {
		kv := strings.Split(kvPair, "=")
		k := kv[0]
		v := kv[1]
		if k == "name" {
			ii.name = v
		} else if k == "order" {
			order32, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return nil, err
			}
			ii.order = int32(order32)
		}
	}

	return ii, nil
}
