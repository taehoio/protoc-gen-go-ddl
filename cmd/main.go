package main

import (
	"flag"
	"fmt"

	"github.com/taehoio/protoc-gen-go-ddl/internal/ddl"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.0.1"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-ddl %v\n", version)
		return
	}

	var flags flag.FlagSet

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		return ddl.GenerateDDLFiles(gen)
	})
}
