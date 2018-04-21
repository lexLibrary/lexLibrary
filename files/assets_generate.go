// +build ignore

package main

import (
	"log"

	"github.com/lexLibrary/lexLibrary/files"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(files.Assets, vfsgen.Options{
		PackageName:  "files",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
