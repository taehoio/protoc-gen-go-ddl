package ddl

import (
	"bytes"
	_ "embed"
	"fmt"
	"path/filepath"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

var (
	//go:embed template/dml_package.pb.go.tmpl
	dmlPackageTmpl string

	//go:embed template/dml_package_mysql.pb.go.tmpl
	dmlPackageMysqlTmpl string

	//go:embed template/dml_package_pagination.pb.go.tmpl
	dmlPackagePaginationTmpl string
)

func generateNecessaryDMLFiles(
	gen *protogen.Plugin,
	version string,
	goPackageName string,
	sourceFile *protogen.File,
	goPackageVisited map[string]struct{},
) error {
	if _, ok := goPackageVisited[goPackageName]; !ok {
		generatedFilenamePrefix := filepath.Join(filepath.Dir(sourceFile.GeneratedFilenamePrefix), goPackageName)

		dmlPackageFileSuffix := "_dml.pb.go"
		generatedDMLPackageFile := gen.NewGeneratedFile(generatedFilenamePrefix+dmlPackageFileSuffix, sourceFile.GoImportPath)
		addFileHead(version, generatedDMLPackageFile, gen, nil, "//")
		generatedDMLPackage, err := generateDMLPackage(goPackageName)
		if err != nil {
			return err
		}
		generatedDMLPackageFile.P(generatedDMLPackage)

		dmlPackageMysqlFileSuffix := "_dml_mysql.pb.go"
		generatedDMLPackageMysqlFile := gen.NewGeneratedFile(generatedFilenamePrefix+dmlPackageMysqlFileSuffix, sourceFile.GoImportPath)
		addFileHead(version, generatedDMLPackageMysqlFile, gen, nil, "//")
		generatedDMLPackageMysql, err := generateDMLPackageMysql(goPackageName)
		if err != nil {
			return err
		}
		generatedDMLPackageMysqlFile.P(generatedDMLPackageMysql)

		dmlPackagePaginationFileSuffix := "_dml_pagination.pb.go"
		generatedDMLPackagePaginationFile := gen.NewGeneratedFile(generatedFilenamePrefix+dmlPackagePaginationFileSuffix, sourceFile.GoImportPath)
		addFileHead(version, generatedDMLPackagePaginationFile, gen, nil, "//")
		generatedDMLPackagePagination, err := generateDMLPackagePagination(goPackageName)
		if err != nil {
			return err
		}
		generatedDMLPackagePaginationFile.P(generatedDMLPackagePagination)
	}
	goPackageVisited[goPackageName] = struct{}{}

	return nil
}

func GenerateDDLFiles(version string, gen *protogen.Plugin) error {
	var (
		goPackageVisited = make(map[string]struct{})
	)

	for _, sourceFile := range gen.Files {
		if !sourceFile.Generate {
			continue
		}

		goPackageName := getGoPackageName(sourceFile.Proto.Options)

		for _, message := range sourceFile.Messages {
			mi, err := NewMessageInfo(*message)
			if err != nil {
				return err
			}

			ddlFileSuffix, err := mi.DDLFileSuffix()
			if err == ErrorDatastoreUnspecified || err == ErrNotSupportedDatastore {
				continue
			}

			if err := generateNecessaryDMLFiles(gen, version, goPackageName, sourceFile, goPackageVisited); err != nil {
				return err
			}

			generatedDDLFile := gen.NewGeneratedFile(sourceFile.GeneratedFilenamePrefix+ddlFileSuffix, sourceFile.GoImportPath)
			switch filepath.Ext(ddlFileSuffix) {
			case ".js":
				addFileHead(version, generatedDDLFile, gen, sourceFile, "//")
			default:
				addFileHead(version, generatedDDLFile, gen, sourceFile, "--")
			}

			stmts, err := mi.GenerateDDL()
			if err == ErrorDatastoreUnspecified || err == ErrNotSupportedDatastore {
				continue
			}
			if err != nil {
				return err
			}
			generatedDDLFile.P(stmts)

			messageDMLFileSuffix, err := mi.DMLFileSuffix()
			if err != nil {
				return err
			}

			messageDMLFilepath := sourceFile.GeneratedFilenamePrefix + messageDMLFileSuffix
			generatedMessageDMLFile := gen.NewGeneratedFile(messageDMLFilepath, sourceFile.GoImportPath)
			addFileHead(version, generatedMessageDMLFile, gen, sourceFile, "//")

			messageDMLMockFileSuffix, err := mi.DMLMockFileSuffix()
			if err != nil {
				return err
			}
			messageDMLMockFilepath := sourceFile.GeneratedFilenamePrefix + messageDMLMockFileSuffix
			d, err := mi.GenerateDML(filepath.Base(messageDMLFilepath), filepath.Base(messageDMLMockFilepath))
			if err != nil {
				return err
			}
			generatedMessageDMLFile.P(d)
		}
	}

	return nil
}

func generateDMLPackage(goPackageName string) (string, error) {
	tmpl, err := template.New("dmlPackageTmpl").Parse(dmlPackageTmpl)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, &dml{
		PackageName: goPackageName,
	}); err != nil {
		return "", err
	}

	return b.String(), nil
}

func generateDMLPackageMysql(goPackageName string) (string, error) {
	tmpl, err := template.New("dmlPackageTmpl").Parse(dmlPackageMysqlTmpl)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, &dml{
		PackageName: goPackageName,
	}); err != nil {
		return "", err
	}

	return b.String(), nil
}

func generateDMLPackagePagination(goPackageName string) (string, error) {
	tmpl, err := template.New("dmlPackagePaginationTmpl").Parse(dmlPackagePaginationTmpl)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, &dml{
		PackageName: goPackageName,
	}); err != nil {
		return "", err
	}

	return b.String(), nil
}

func addFileHead(version string, g *protogen.GeneratedFile, gen *protogen.Plugin, sourceFile *protogen.File, lineComment string) {
	g.P(fmt.Sprintf("%s Code generated by protoc-gen-go-ddl. DO NOT EDIT.", lineComment))
	g.P(fmt.Sprintf("%s versions:", lineComment))
	g.P(fmt.Sprintf("%s  protoc-gen-go-ddl v%s", lineComment, version))
	g.P(fmt.Sprintf("%s  protoc            %s", lineComment, protocVersion(gen)))
	if sourceFile != nil {
		if sourceFile.Proto.GetOptions().GetDeprecated() {
			g.P(fmt.Sprintf("%s %s is a deprecated file.", lineComment, sourceFile.Desc.Path()))
		} else {
			g.P(fmt.Sprintf("%s source: %s", lineComment, sourceFile.Desc.Path()))
		}
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
