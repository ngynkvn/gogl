package main

import (
	"fmt"
	shaders "game/gen"
	"math"
	"os"
	"runtime"
	"strings"
	"time"
	"unsafe"

	gl "github.com/go-gl/gl/v4.1-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

//go:generate go run codegen/gen_shaders.go shaders/fragment/basic.fragment.shader BasicFragmentShader
//go:generate go run codegen/gen_shaders.go shaders/vertex/basic.vertex.shader BasicVertexShader

type Shader = uint32

// Use this as a sanity check on opengl context
func printversion() {
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
}

func CreateShader(shaderType Shader, source string) (Shader, error) {
	s := gl.CreateShader(shaderType)
	s_src, free := gl.Strs(source)
	gl.ShaderSource(s, 1, s_src, nil)
	free()
	gl.CompileShader(s)
	var status int32
	gl.GetShaderiv(s, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &status)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(s, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to compile source %v:\n=====\n%v====", source, log)
	} else if status == gl.TRUE {
		fmt.Printf("Compiled Shader✅: %v\n", s)
	}
	return Shader(s), nil
}

func createprogram() uint32 {
	// VERTEX SHADER
	vs, err := CreateShader(gl.VERTEX_SHADER, shaders.BasicVertexShader)
	if err != nil {
		fmt.Println(err)
	}

	// FRAGMENT SHADER
	fs, err := CreateShader(gl.FRAGMENT_SHADER, shaders.BasicFragmentShader)
	if err != nil {
		fmt.Println(err)
	}

	// CREATE PROGRAM
	program := gl.CreateProgram()
	gl.AttachShader(program, vs)
	gl.AttachShader(program, fs)
	fragoutstring := gl.Str("outColor\x00")
	gl.BindFragDataLocation(program, 0, fragoutstring)

	gl.LinkProgram(program)
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		println(fmt.Errorf("failed to compile:\n %v\n++++++++", log))
	}

	return program
}

var uniRoll float32 = 0.0
var uniYaw float32 = 1.0
var uniPitch float32 = 0.0
var uniscale float32 = 0.3
var yrot float32 = 20.0
var zrot float32 = 0.0
var xrot float32 = 0.0
var UniScale int32

func glDebugCallback(
	source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer) {
	fmt.Fprintf(
		os.Stderr,
		"Debug (source: %d, type: %d severity: %d): %s\n",
		source, gltype, severity, message)
}

func main() {
	var window *sdl.Window
	var context sdl.GLContext
	var event sdl.Event
	var running bool
	var err error
	runtime.LockOSThread()
	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_OPENGL)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	context, err = window.GLCreateContext()
	if err != nil {
		panic(err)
	}
	defer sdl.GLDeleteContext(context)

	gl.Init()
	gl.Enable(gl.DEBUG_OUTPUT)

	gl.DebugMessageCallback(glDebugCallback, nil)
	gl.Viewport(0, 0, winWidth, winHeight)
	// OPENGL FLAGS
	gl.ClearColor(0.0, 0.1, 0.0, 1.0)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// VERTEX BUFFER
	var vertexbuffer uint32
	gl.GenBuffers(1, &vertexbuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexbuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(triangle_vertices)*4, gl.Ptr(triangle_vertices), gl.STATIC_DRAW)

	// COLOUR BUFFER
	var colourbuffer uint32
	gl.GenBuffers(1, &colourbuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(triangle_colours)*4, gl.Ptr(triangle_colours), gl.STATIC_DRAW)

	// GUESS WHAT
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
	program := createprogram()

	// VERTEX ARRAY
	var VertexArrayID uint32
	gl.GenVertexArrays(1, &VertexArrayID)
	gl.BindVertexArray(VertexArrayID)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexbuffer)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	// VERTEX ARRAY HOOK COLOURS
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil)

	//UNIFORM HOOK
	unistring := gl.Str("scaleMove\x00")
	UniScale = gl.GetUniformLocation(program, unistring)
	fmt.Printf("Uniform Link: %v\n", UniScale+1)

	gl.UseProgram(program)

	running = true
	for running {
		for event = sdl.PollEvent(); event != nil; event =
			sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:

				xrot = float32(t.Y) / 2
				yrot = float32(t.X) / 2
				fmt.Printf("[%dms]MouseMotion\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Which, t.X, t.Y, t.XRel, t.YRel)
			}
		}
		drawgl()
		window.GLSwap()
	}
}

func drawgl() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	uniYaw = yrot * (math.Pi / 180.0)
	yrot = yrot - 1.0
	uniPitch = zrot * (math.Pi / 180.0)
	zrot = zrot - 0.5
	uniRoll = xrot * (math.Pi / 180.0)
	xrot = xrot - 0.2

	gl.Uniform4f(UniScale, uniRoll, uniYaw, uniPitch, uniscale)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangle_vertices)*4))

	time.Sleep(50 * time.Millisecond)

}

const (
	winTitle  = "OpenGL Shader"
	winWidth  = 640
	winHeight = 480
)

var triangle_vertices = []float32{
	-1.0, -1.0, -1.0,
	-1.0, -1.0, 1.0,
	-1.0, 1.0, 1.0,
	1.0, 1.0, -1.0,
	-1.0, -1.0, -1.0,
	-1.0, 1.0, -1.0,
	1.0, -1.0, 1.0,
	-1.0, -1.0, -1.0,
	1.0, -1.0, -1.0,
	1.0, 1.0, -1.0,
	1.0, -1.0, -1.0,
	-1.0, -1.0, -1.0,
	-1.0, -1.0, -1.0,
	-1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0,
	1.0, -1.0, 1.0,
	-1.0, -1.0, 1.0,
	-1.0, -1.0, -1.0,
	-1.0, 1.0, 1.0,
	-1.0, -1.0, 1.0,
	1.0, -1.0, 1.0,
	1.0, 1.0, 1.0,
	1.0, -1.0, -1.0,
	1.0, 1.0, -1.0,
	1.0, -1.0, -1.0,
	1.0, 1.0, 1.0,
	1.0, -1.0, 1.0,
	1.0, 1.0, 1.0,
	1.0, 1.0, -1.0,
	-1.0, 1.0, -1.0,
	1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0,
	-1.0, 1.0, 1.0,
	1.0, 1.0, 1.0,
	-1.0, 1.0, 1.0,
	1.0, -1.0, 1.0}

var triangle_colours = []float32{
	0.583, 0.771, 0.014,
	0.609, 0.115, 0.436,
	0.327, 0.483, 0.844,
	0.822, 0.569, 0.201,
	0.435, 0.602, 0.223,
	0.310, 0.747, 0.185,
	0.597, 0.770, 0.761,
	0.559, 0.436, 0.730,
	0.359, 0.583, 0.152,
	0.483, 0.596, 0.789,
	0.559, 0.861, 0.639,
	0.195, 0.548, 0.859,
	0.014, 0.184, 0.576,
	0.771, 0.328, 0.970,
	0.406, 0.615, 0.116,
	0.676, 0.977, 0.133,
	0.971, 0.572, 0.833,
	0.140, 0.616, 0.489,
	0.997, 0.513, 0.064,
	0.945, 0.719, 0.592,
	0.543, 0.021, 0.978,
	0.279, 0.317, 0.505,
	0.167, 0.620, 0.077,
	0.347, 0.857, 0.137,
	0.055, 0.953, 0.042,
	0.714, 0.505, 0.345,
	0.783, 0.290, 0.734,
	0.722, 0.645, 0.174,
	0.302, 0.455, 0.848,
	0.225, 0.587, 0.040,
	0.517, 0.713, 0.338,
	0.053, 0.959, 0.120,
	0.393, 0.621, 0.362,
	0.673, 0.211, 0.457,
	0.820, 0.883, 0.371,
	0.982, 0.099, 0.879}
