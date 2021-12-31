package ddl

import (
	"path"

	"google.golang.org/protobuf/compiler/protogen"
)

func removeExt(filename, ext string) string {
	if path.Ext(filename) == ext {
		return filename[:len(filename)-len(ext)]
	}
	return filename
}

func generatingFilePath(sourceFilePath string) string {
	generatingFilePath := removeExt(sourceFilePath, ".proto")
	generatingFilePath += "_ddl.pb.sql"
	return generatingFilePath
}

func GenerateDDLFiles(gen *protogen.Plugin) error {
	extensionTypes, err := loadAllExtensionTypes(*gen)
	if err != nil {
		return err
	}

	for _, sourceFile := range gen.Files {
		if !sourceFile.Generate {
			continue
		}

		generatedFile := gen.NewGeneratedFile(generatingFilePath(sourceFile.Desc.Path()), protogen.GoImportPath(sourceFile.GoPackageName))

		for _, message := range sourceFile.Messages {
			mi, err := NewMessageInfo(*message, extensionTypes)
			if err != nil {
				return err
			}

			stmts, err := mi.GenerateDDLToCreate()
			if err != nil {
				return err
			}

			generatedFile.P(stmts)
		}
	}

	return nil
}
