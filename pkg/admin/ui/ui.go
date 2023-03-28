package ui

import "embed"

// StaticFiles includes all ui app static contents
//
//go:embed dist/*
var StaticFiles embed.FS
