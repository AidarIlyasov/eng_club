package handlers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/valyala/fasthttp"
)

// ServeFrontend serves the frontend static files
func (d *Deps) ServeFrontend(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	
	// Remove leading slash
	if path == "/" {
		path = "/index.html"
	}
	
	// Build file path
	frontendDir := "./frontend/dist"
	filePath := filepath.Join(frontendDir, path)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// For SPA routing, serve index.html for non-API routes
		if !strings.HasPrefix(path, "/api/") {
			filePath = filepath.Join(frontendDir, "index.html")
		} else {
			ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
			ctx.Response.SetBodyString("Not found")
			return
		}
	}
	
	// Serve the file
	fasthttp.ServeFile(ctx, filePath)
}