package middleware

import (
    "log"
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

    log.Printf("✅ 找到前端构建文件: %s", distPath)

    r.Static("/static", filepath.Join(distPath, "static"))
    r.Static("/js", filepath.Join(distPath, "js"))
    r.Static("/css", filepath.Join(distPath, "css"))
    r.StaticFile("/favicon.ico", filepath.Join(distPath, "favicon.ico"))
    
    r.NoRoute(func(c *gin.Context) {
        path := c.Request.URL.Path
        
        if len(path) >= 4 && path[:4] == "/api" {
            c.JSON(404, gin.H{
                "code":    404,
                "message": "API endpoint not found",
                "path":    path,
            })
            return
        }
        
        if strings.Contains(path, ".") {
            c.Status(404)
            return
        }
        
        indexPath := filepath.Join(distPath, "index.html")
        log.Printf("🏠 返回前端页面: %s -> %s", path, indexPath)
        c.File(indexPath)
    })
}
