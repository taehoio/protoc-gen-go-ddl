package ddl

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Recursively register all extensions into the provided protoregistry.Types,
// starting with the protoreflect.FileDescriptor and recursing into its MessageDescriptors,
// their nested MessageDescriptors, and so on.
//
// This leverages the fact that both protoreflect.FileDescriptor and protoreflect.MessageDescriptor
// have identical Messages() and Extensions() functions in order to recurse through a single function
func registerAllExtensions(extTypes *protoregistry.Types, descs interface {
	Messages() protoreflect.MessageDescriptors
	Extensions() protoreflect.ExtensionDescriptors
}) error {
	mds := descs.Messages()
	for i := 0; i < mds.Len(); i++ {
		if err := registerAllExtensions(extTypes, mds.Get(i)); err != nil {
			return err
		}
	}
	xds := descs.Extensions()
	for i := 0; i < xds.Len(); i++ {
		if err := extTypes.RegisterExtension(dynamicpb.NewExtensionType(xds.Get(i))); err != nil {
			return err
		}
	}
	return nil
}

func loadAllExtensionTypes(gen protogen.Plugin) (protoregistry.Types, error) {
	var extTypes protoregistry.Types

	for _, file := range gen.Files {
		if err := registerAllExtensions(&extTypes, file.Desc); err != nil {
			return extTypes, err
		}
	}

	return extTypes, nil
}
