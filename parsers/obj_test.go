package parser

import "testing"

func Test(t *testing.T) {
	objParser := NewObjParser()
	objParser.Parse("../assets/bunny.obj")
}
