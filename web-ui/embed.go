package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:build/*
var static embed.FS

var FS, _ = fs.Sub(static, "build")
