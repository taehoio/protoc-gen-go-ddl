package main

import (
	"flag"
	"fmt"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

const version = "0.0.1-alpha"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-marshal-input %v\n", version)
		return
	}

	var flags flag.FlagSet

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		opt := proto.MarshalOptions{Deterministic: true}
		b, err := opt.Marshal(gen.Request)
		if err != nil {
			return err
		}

		if err := os.WriteFile("marshaled_input.dat", b, 0644); err != nil {
			return err
		}

		return nil
	})
}
