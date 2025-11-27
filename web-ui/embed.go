package webui

import (
	"embed"
	"io/fs"
)

//go:embed build/*
var static embed.FS

var FS, _ = fs.Sub(static, "build")
