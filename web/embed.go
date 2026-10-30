package web

import "embed"

// FS holds all embedded web application assets (HTML, CSS, JS, fonts).
//
//go:embed index.html css/* js/* assets/fonts/*
var FS embed.FS
