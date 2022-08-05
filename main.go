package main

import (
	"fmt"
	shaders "game/gen"
	parser "game/parsers"

	"game/log"
	"runtime"
	"strings"
	"time"

	"github.com/chewxy/math32"
	"github.com/engoengine/glm"
	gl "github.com/go-gl/gl/v4.1-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

//go:generate go run codegen/gen_shaders.go shaders/fragment/basic.fragment.shader BasicFragmentShader
//go:generate go run codegen/gen_shaders.go shaders/vertex/basic.vertex.shader BasicVertexShader

type Shader = uint32

var logger = log.Logger()

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
		logger.Debugf("Compiled Shader✅: %v\n", s)
	}
	return Shader(s), nil
}

func createprogram() uint32 {
	// VERTEX SHADER
	vs, err := CreateShader(gl.VERTEX_SHADER, shaders.BasicVertexShader)
	if err != nil {
		logger.Debugf(`(source)
		%v
		==================
		(err reported)
		%v
		`, shaders.BasicVertexShader, err)
	}

	// FRAGMENT SHADER
	fs, err := CreateShader(gl.FRAGMENT_SHADER, shaders.BasicFragmentShader)
	if err != nil {
		logger.Debugf(`(source)
		%v
		==================
		(err reported)
		%v
		`, shaders.BasicFragmentShader, err)
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

		lg := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(lg))

		logger.Debugf("%v", fmt.Errorf("failed to compile:\n %v\n++++++++", lg))
	}

	return program
}

var UniScale int32
var Model int32
var View int32
var Projection int32

var cameraPos = glm.Vec3{0.0, 0.0, 3.0}
var cameraFront = glm.Vec3{0.0, 0.0, -1.0}
var cameraUp = glm.Vec3{0.0, 1.0, 0.0}
var dx float32 = 0.0
var dy float32 = 0.0

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
	sdl.SetRelativeMouseMode(true)
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

	// Oh god
	objParser := parser.NewObjParser()
	parsed := objParser.Parse("assets/b.obj")
	vertices := make([]float32, 0)
	for _, v := range parsed.Faces {
		v1, v2, v3 := parsed.Vertices[v[0]-1], parsed.Vertices[v[1]-1], parsed.Vertices[v[2]-1]
		vertices = append(vertices, v1[:]...)
		vertices = append(vertices, v2[:]...)
		vertices = append(vertices, v3[:]...)
	}

	// VERTEX BUFFER
	var vertexbuffer uint32
	gl.GenBuffers(1, &vertexbuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexbuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// COLOUR BUFFER
	var colourbuffer uint32
	gl.GenBuffers(1, &colourbuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

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

	gl.UseProgram(program)

	//UNIFORM HOOK
	sm := gl.Str("model\x00")
	Model = gl.GetUniformLocation(program, sm)
	logger.Debugf("[*]Uniform location fetched: %v\n", Model)

	v := gl.Str("view\x00")
	View = gl.GetUniformLocation(program, v)
	logger.Debugf("[*]Uniform location fetched: %v\n", View)

	p := gl.Str("projection\x00")
	Projection = gl.GetUniformLocation(program, p)
	logger.Debugf("[*]Uniform location fetched: %v\n", Projection)

	running = true
	for running {
		for event = sdl.PollEvent(); event != nil; event =
			sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				fmt.Printf("[%dms]MouseMotion\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Which, t.X, t.Y, t.XRel, t.YRel)
				dx = float32(t.XRel)
				dy = float32(t.YRel)
			case *sdl.KeyboardEvent:
				switch t.Keysym.Sym {
				case sdl.K_w:
					v := cameraFront.Mul(0.1)
					cameraPos = cameraPos.Add(&v)
				case sdl.K_s:
					v := cameraFront.Mul(0.1)
					cameraPos = cameraPos.Sub(&v)
				case sdl.K_a:
					nv := glm.NormalizeVec3(cameraFront.Cross(&cameraUp))
					v := nv.Mul(0.1)
					cameraPos = cameraPos.Sub(&v)
				case sdl.K_d:
					nv := glm.NormalizeVec3(cameraFront.Cross(&cameraUp))
					v := nv.Mul(0.1)
					cameraPos = cameraPos.Add(&v)
				}
			}
		}
		drawgl(len(vertices))
		dx = 0
		dy = 0
		window.GLSwap()
	}
}

func CheckGLErrors() {
	errCode := gl.GetError()
	if errCode == gl.NO_ERROR {
		return
	}

}

var yaw float32 = 0.0
var pitch float32 = 0.0

func drawgl(n int) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	yaw += dx
	pitch += dy
	if pitch > 89.0 {
		pitch = 89.0
	}
	if pitch < -89.0 {
		pitch = -89.0
	}

	direction := glm.Vec3{}
	direction[0] = math32.Cos(glm.DegToRad(yaw)) * math32.Cos(glm.DegToRad(pitch))
	direction[1] = math32.Sin(glm.DegToRad(pitch))
	direction[2] = math32.Sin(glm.DegToRad(yaw)) * math32.Cos(glm.DegToRad(pitch))
	cameraFront = glm.NormalizeVec3(direction)

	view := glm.LookAtV(&cameraPos, &cameraFront, &cameraUp)
	gl.UniformMatrix4fv(View, 1, false, &view[0])

	projection := glm.Perspective(glm.DegToRad(45.0), float32(winWidth)/float32(winHeight), 0.1, 100.0)
	gl.UniformMatrix4fv(Projection, 1, false, &projection[0])

	model := glm.Ident4()
	gl.UniformMatrix4fv(Model, 1, false, &model[0])

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(n*4))

	time.Sleep(16 * time.Millisecond)

}

const (
	winTitle  = "OpenGL Shader"
	winWidth  = 640
	winHeight = 480
)
