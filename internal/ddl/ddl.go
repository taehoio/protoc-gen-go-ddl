package ddl

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

var (
	//go:embed template/dml.pb.go.tmpl
	dmlTmpl string
)

func GenerateDDLFiles(version string, gen *protogen.Plugin) error {
	for _, sourceFile := range gen.Files {
		if !sourceFile.Generate {
			continue
		}

		dmlFileSuffix := "_dml.pb.go"
		generatedDMLFile := gen.NewGeneratedFile(sourceFile.GeneratedFilenamePrefix+dmlFileSuffix, sourceFile.GoImportPath)
		addGoFileHead(version, generatedDMLFile, gen, sourceFile)
		generatedDML, err := generateDML(sourceFile)
		if err != nil {
			return err
		}
		generatedDMLFile.P(generatedDML)

		for _, message := range sourceFile.Messages {
			mi, err := NewMessageInfo(*message)
			if err != nil {
				return err
			}

			messageName := strcase.ToSnake(string(message.Desc.Name()))

			sqlFileSuffix := fmt.Sprintf("_ddl_%s.pb.sql", messageName)
			generatedSQLFile := gen.NewGeneratedFile(sourceFile.GeneratedFilenamePrefix+sqlFileSuffix, sourceFile.GoImportPath)
			addSQLFileHead(version, generatedSQLFile, gen, sourceFile)
			stmts, err := mi.GenerateDDL()
			if err == ErrNotSupportedDatastore {
				continue
			}
			if err != nil {
				return err
			}
			generatedSQLFile.P(stmts)

			messageDMLFileSuffix := fmt.Sprintf("_dml_%s.pb.go", messageName)
			gereratedMessageDMLFile := gen.NewGeneratedFile(sourceFile.GeneratedFilenamePrefix+messageDMLFileSuffix, sourceFile.GoImportPath)
			addGoFileHead(version, gereratedMessageDMLFile, gen, sourceFile)
			d, err := mi.GenerateDML()
			if err != nil {
				return err
			}
			gereratedMessageDMLFile.P(d)
		}
	}

	return nil
}

func generateDML(sourceFile *protogen.File) (string, error) {
	tmpl, err := template.New("dmlTmpl").Parse(dmlTmpl)
	if err != nil {
		return "", err
	}

	pkgName := getGoPackageName(sourceFile.Proto.Options)
	var b bytes.Buffer
	if err := tmpl.Execute(&b, &dml{
		PackageName: pkgName,
	}); err != nil {
		return "", err
	}

	return b.String(), nil
}

func addSQLFileHead(version string, g *protogen.GeneratedFile, gen *protogen.Plugin, sourceFile *protogen.File) {
	g.P("-- Code generated by protoc-gen-go-ddl. DO NOT EDIT.")
	g.P("-- versions:")
	g.P("--  protoc-gen-go-ddl v", version)
	g.P("--  protoc            ", protocVersion(gen))
	if sourceFile.Proto.GetOptions().GetDeprecated() {
		g.P("-- ", sourceFile.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("-- source: ", sourceFile.Desc.Path())
	}
}

func addGoFileHead(version string, g *protogen.GeneratedFile, gen *protogen.Plugin, sourceFile *protogen.File) {
	g.P("// Code generated by protoc-gen-go-ddl. DO NOT EDIT.")
	g.P("// versions:")
	g.P("//  protoc-gen-go-ddl v", version)
	g.P("//  protoc            ", protocVersion(gen))
	if sourceFile.Proto.GetOptions().GetDeprecated() {
		g.P("// ", sourceFile.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", sourceFile.Desc.Path())
	}
	g.P()
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}
