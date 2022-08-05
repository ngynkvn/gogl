package main

import (
	"fmt"
	"unsafe"

	gl "github.com/go-gl/gl/v4.1-core/gl"
)

func glDebugCallback(
	source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer) {
	logger.Infow("[DEBUG]", "source", source, "gltype", gltype, "id", id, "severity", severity, "message", message)
}

// Use this as a sanity check on opengl context
func printversion() {
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
}
