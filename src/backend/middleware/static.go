package middleware

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func ServeStaticFiles(r *gin.Engine) {
	distPath := "../frontend/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		log.Println("⚠️  前端构建文件不存在，跳过静态文件服务")
		log.Println("请运行: cd frontend && npm run build")
		return
	}

	log.Printf("✅ 找到前端构建文件: %s, 将提供静态文件服务", distPath)

	// Serve static files from specific paths to avoid conflicts with API routes
	r.Static("/js", filepath.Join(distPath, "js"))
	r.Static("/css", filepath.Join(distPath, "css"))
	r.Static("/img", filepath.Join(distPath, "img"))
	r.Static("/fonts", filepath.Join(distPath, "fonts"))

	// Serve favicon and other root files individually
	r.StaticFile("/favicon.ico", filepath.Join(distPath, "favicon.ico"))
	r.StaticFile("/manifest.json", filepath.Join(distPath, "manifest.json"))

	// Custom NoRoute handler to redirect all non-API 404 GET requests to the SPA entry point.
	r.NoRoute(func(c *gin.Context) {
		// Only handle GET requests for the SPA fallback.
		if c.Request.Method != "GET" {
			c.Status(http.StatusNotFound)
			return
		}

		path := c.Request.URL.Path

		// If it's an API call, let it 404 as it's a genuinely unknown endpoint.
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "API endpoint not found"})
			return
		}

		// For all other GET requests that didn't match a static file,
		// it's a frontend route. Serve the SPA's entry point.
		log.Printf("🏠 路由未匹配 '%s', 返回前端主页面", path)
		c.File(filepath.Join(distPath, "index.html"))
	})
}
