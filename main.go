// Genbuilder is a code generator for implementing a builder along with providing getters and setters
// for non public members of the given struct.

// Usage genbuilder /path/to/struct/definition
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

/*
Ready to start creating getters and the builder
*/
const path = "/Users/dsearle/go/src/playground/builders"

var pathSeparator = ""

type Jason struct {
	t         time.Time
	boolean   bool
	floater   float64
	reader    io.Reader
	integer   int
	integer32 int32
	integer64 int64
}

func main() {
	pathSeparator = fmt.Sprintf("%c", os.PathSeparator)
	seedType := reflect.TypeOf(Jason{})
	genBuilder(seedType)
	genImmutable(seedType)
}

func genBuilder(seedType reflect.Type) {
	fileName := toPrivate(seedType.Name())
	file, fpath := createFile(path, fileName+"builder")
	defer file.Close()
	writePackage(file, path)
	structName := writeBuilderStruct(file, seedType)
	writeConstructor(file, structName)
	numFields := seedType.NumField()
	for i := 0; i < numFields; i++ {
		writeSetter(file, structName, seedType.Field(i))
	}
	writeBuildMethod(file, structName, seedType)
	c := exec.Command("goimports", "-w", fpath)
	c.Run()
}

func genImmutable(seedType reflect.Type) {
	fname := toPrivate(seedType.Name())
	file, fpath := createFile(path, fname)
	defer file.Close()
	writePackage(file, path)
	structName := writeStruct(file, seedType)
	writeConstructor(file, structName)
	numFields := seedType.NumField()
	for i := 0; i < numFields; i++ {
		writeGetter(file, structName, seedType.Field(i))
	}
	writeAsBuilderMethod(file, seedType)
	c := exec.Command("goimports", "-w", fpath)
	c.Run()
}

func createFile(path, fname string) (*os.File, string) {
	fpath := buildDirs(path, fname)
	file, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return file, fpath
}

func buildDirs(path, fileName string) string {
	p := pathSeparator
	dirs := strings.Split(path, pathSeparator)
	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		p = filepath.Join(p, dir)
		fmt.Println(p)
		checkDir(p)
	}
	return filepath.Join(p, fileName+".go")
}

func checkDir(dir string) {
	_, err := os.Open(dir)
	if err != nil {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			fmt.Println("unable to create directory ", dir)
			os.Exit(1)
		}
	}
}

func writePackage(file *os.File, path string) {
	dirs := strings.Split(path, pathSeparator)
	pkg := dirs[len(dirs)-1]
	file.WriteString("package " + pkg + "\n")
}

func writeConstructor(file *os.File, structType string) {
	file.WriteString("\nfunc New" + toPublic(structType) + "() *" + structType + " {\n")
	file.WriteString("return new(" + structType + ") \n}\n")
}

func writeBuilderStruct(file *os.File, seedType reflect.Type) string {
	structType := toPublic(seedType.Name()) + "Builder"
	file.WriteString("type " + structType + " struct{\n")
	numFields := seedType.NumField()
	for i := 0; i < numFields; i++ {
		field := seedType.Field(i)
		file.WriteString(toPublic(field.Name) + " " + field.Type.String() + "\n")
	}
	file.WriteString("}")
	return structType
}

func writeStruct(file *os.File, seedType reflect.Type) string {
	structType := seedType.Name()
	file.WriteString("type " + structType + " struct{\n")
	numFields := seedType.NumField()
	for i := 0; i < numFields; i++ {
		field := seedType.Field(i)
		file.WriteString(field.Name + " " + field.Type.String() + "\n")
	}
	file.WriteString("}\n")
	return structType
}

func writeGetter(file *os.File, structName string, field reflect.StructField) {
	rcvr := strings.ToLower(string([]rune(structName)[0]))
	fxName := toPublic(field.Name)
	file.WriteString("\nfunc ( " + rcvr + " *" + structName + ") " + fxName + "()" + field.Type.String() + " {\n")
	file.WriteString("return " + rcvr + "." + field.Name)
	file.WriteString("\n}\n")
}

func writeSetter(file *os.File, structName string, field reflect.StructField) {
	rcvr := strings.ToLower(string([]rune(structName)[0]))
	fxName := toPublic(field.Name)
	file.WriteString("\nfunc ( " + rcvr + " *" + structName + ") Set" + fxName + "(" + field.Name + " " + field.Type.String() + ") *" + structName + " {\n")
	file.WriteString(rcvr + "." + fxName + "=" + field.Name)
	file.WriteString("\nreturn " + rcvr)
	file.WriteString("\n}\n")
}

func writeAsBuilderMethod(file *os.File, structName string, seedType reflect.Type) {
	builder := structName + "Builder"
	rcvr := strings.ToLower(string([]rune(structName)[0]))
	file.WriteString("\nfunc ( " + rcvr + " *" + structName + ") AsBuilder() *" + builder + " {\n")
	file.WriteString("return &" + builder + "{\n")
	numFields := seedType.NumField()
	for i := 0; i < numFields; i++ {
		field := seedType.Field(i)
		file.WriteString(toPublic(field.Name) + ": " + rcvr + "." + field.Name + ",\n")
	}
	file.WriteString("}\n}\n")
}

func writeBuildMethod(file *os.File, structName string, seedType reflect.Type) {
	rcvr := strings.ToLower(string([]rune(structName)[0]))
	file.WriteString("\nfunc ( " + rcvr + " *" + structName + ") Build () *" + seedType.Name() + " {\n")
	file.WriteString("return &" + seedType.Name() + "{\n")
	numFields := seedType.NumField()
	for i := 0; i < numFields; i++ {
		f := seedType.Field(i)
		file.WriteString(f.Name + ": " + rcvr + "." + toPublic(f.Name+",\n"))
	}
	file.WriteString("}")
	file.WriteString("\n}\n")
}

func toPublic(str string) string {
	runes := []rune(str)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

func toPrivate(str string) string {
	runes := []rune(str)
	runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
	return string(runes)
}
