package parser

import (
	"bufio"
	"game/log"
	"os"
	"strconv"
	"strings"
)

var logger = log.Logger()

type Vertex [3]float32
type Index [3]uint32
type Obj struct {
	Vertices []Vertex
	Faces    []Index
}

type ObjParser struct {
	lineCount uint32
	Obj       Obj
}

type LineType uint

const (
	V LineType = iota
	F
	U
	C
)

func Map[T any, K any](input []T, fn func(t T) K) []K {
	out := make([]K, len(input))
	for i, v := range input {
		out[i] = fn(v)
	}
	return out
}

func convF32(f string) float32 {
	r, e := strconv.ParseFloat(f, 32)
	if e != nil {
		logger.Fatal(e)
	}
	return float32(r)
}
func convU32(f string) uint32 {
	r, e := strconv.ParseUint(f, 10, 32)
	if e != nil {
		logger.Fatal(e)
	}
	return uint32(r)
}

func (obj *ObjParser) parseline(txt string) LineType {
	txtArr := strings.Split(txt, " ")
	switch txtArr[0] {
	case "v":
		values := Map(txtArr[1:], convF32)
		obj.Obj.Vertices = append(obj.Obj.Vertices, Vertex{values[0], values[1], values[2]})
		return LineType(V)
	case "f":
		values := Map(txtArr[1:], convU32)
		obj.Obj.Faces = append(obj.Obj.Faces, Index{values[0], values[1], values[2]})
		return LineType(F)
	case "#":
		return LineType(C)
	}
	logger.Fatalf("Yo wtf is %v", txt)
	panic(69)
}

func (obj *ObjParser) Parse(filepath string) Obj {
	readFile, err := os.Open(filepath)
	if err != nil {
		logger.Fatal(err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		switch obj.parseline(fileScanner.Text()) {
		}
		obj.lineCount += 1
	}
	err = readFile.Close()
	if err != nil {
		logger.Fatal(err)
	}
	return obj.Obj
}

func NewObjParser() ObjParser {
	return ObjParser{}
}
