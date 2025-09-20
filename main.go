package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ConfigFile struct {
	Port            string `json:"port"`
	WorkDir         string `json:"workdir"`
	UploadDir       string `json:"uploaddir"`
	FileExpiryHours int    `json:"file_expiry_hours"`
}

func loadConfig(filename string) (*ConfigFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config ConfigFile
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

type Config struct {
	Port            string
	WorkDir         string
	UploadDir       string
	ConfigFile      string
	FileExpiryHours int
}

func main() {
	var config Config

	// 命令行参数
	flag.StringVar(&config.Port, "port", "8080", "服务器端口")
	flag.StringVar(&config.WorkDir, "workdir", "./uploads", "工作目录")
	flag.StringVar(&config.UploadDir, "uploaddir", "./uploads", "上传目录")
	flag.StringVar(&config.ConfigFile, "config", "./config.json", "配置文件路径")
	flag.IntVar(&config.FileExpiryHours, "expiry", 2, "文件过期时间（小时）")
	flag.Parse()

	// 读取配置文件（如果指定）
	if config.ConfigFile != "" {
		cfg, err := loadConfig(config.ConfigFile)
		if err != nil {
			log.Printf("加载配置文件失败: %v", err)
		} else {
			// 配置文件中的值优先级更高
			if cfg.Port != "" {
				config.Port = cfg.Port
			}
			if cfg.WorkDir != "" {
				config.WorkDir = cfg.WorkDir
			}
			if cfg.UploadDir != "" {
				config.UploadDir = cfg.UploadDir
			}
			if cfg.FileExpiryHours > 0 {
				config.FileExpiryHours = cfg.FileExpiryHours
			}
		}
	}

	// 确保工作目录和上传目录存在
	if err := ensureDir(config.WorkDir); err != nil {
		log.Fatalf("创建工作目录失败: %v", err)
	}
	if err := ensureDir(config.UploadDir); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// 转换为绝对路径
	absWorkDir, err := filepath.Abs(config.WorkDir)
	if err != nil {
		log.Fatalf("获取工作目录绝对路径失败: %v", err)
	}
	absUploadDir, err := filepath.Abs(config.UploadDir)
	if err != nil {
		log.Fatalf("获取上传目录绝对路径失败: %v", err)
	}

	config.WorkDir = absWorkDir
	config.UploadDir = absUploadDir

	// 启动文件清理定时任务
	go startFileCleanupTask(config.WorkDir, config.FileExpiryHours)

	// 设置路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveFileBrowser(w, r, config.WorkDir, config.FileExpiryHours)
	})

	http.HandleFunc("/uploads", func(w http.ResponseWriter, r *http.Request) {
		handleUpload(w, r, config.UploadDir)
	})

	addr := ":" + config.Port
	fmt.Printf("文件服务器启动成功！\n")
	fmt.Printf("工作目录: %s\n", config.WorkDir)
	fmt.Printf("上传目录: %s\n", config.UploadDir)
	fmt.Printf("文件过期时间: %d 小时\n", config.FileExpiryHours)
	fmt.Printf("访问地址: http://localhost%s\n", addr)
	fmt.Printf("文件浏览: http://localhost%s/\n", addr)
	fmt.Printf("文件上传: http://localhost%s/uploads\n", addr)

	log.Fatal(http.ListenAndServe(addr, nil))
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func serveFileBrowser(w http.ResponseWriter, r *http.Request, workDir string, expiryHours int) {
	// 安全检查：防止路径遍历
	requestPath := filepath.Clean(r.URL.Path)
	if strings.Contains(requestPath, "..") {
		http.Error(w, "路径不合法", http.StatusBadRequest)
		return
	}

	// 构建完整路径
	fullPath := filepath.Join(workDir, requestPath)

	// 检查文件是否存在
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		http.Error(w, "文件或目录不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		return
	}

	// 如果是目录，显示目录内容
	if info.IsDir() {
		serveDirectory(w, r, fullPath, requestPath, expiryHours)
		return
	}

	// 如果是文件，直接提供下载
	http.ServeFile(w, r, fullPath)
}

func serveDirectory(w http.ResponseWriter, r *http.Request, dirPath, requestPath string, expiryHours int) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "读取目录失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 生成HTML页面
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>文件浏览器 - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:hover { background-color: #f5f5f5; }
        a { text-decoration: none; color: #007bff; }
        a:hover { text-decoration: underline; }
        .upload-section { margin-top: 30px; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        input[type="file"] { margin: 10px 0; }
        input[type="submit"] { background-color: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        input[type="submit"]:hover { background-color: #0056b3; }
        .expired { color: #dc3545; font-weight: bold; }
        .expiring-soon { color: #fd7e14; font-weight: bold; }
        .countdown { font-family: monospace; }
    </style>
    <script>
        function updateCountdowns() {
            const countdowns = document.querySelectorAll('.countdown');
            countdowns.forEach(function(element) {
                const expireTime = parseInt(element.dataset.expireTime);
                const now = Math.floor(Date.now() / 1000);
                const remaining = expireTime - now;
                
                if (remaining <= 0) {
                    element.textContent = '已过期';
                    element.className = 'countdown expired';
                } else {
                    const hours = Math.floor(remaining / 3600);
                    const minutes = Math.floor((remaining %% 3600) / 60);
                    const seconds = remaining %% 60;
                    
                    let timeStr = '';
                    if (hours > 0) {
                        timeStr = hours + '时' + minutes + '分' + seconds + '秒';
                    } else if (minutes > 0) {
                        timeStr = minutes + '分' + seconds + '秒';
                    } else {
                        timeStr = seconds + '秒';
                    }
                    
                    element.textContent = timeStr;
                    
                    if (remaining < 3600) { // 小于1小时
                        element.className = 'countdown expiring-soon';
                    } else {
                        element.className = 'countdown';
                    }
                }
            });
        }
        
        // 页面加载完成后开始倒计时
        document.addEventListener('DOMContentLoaded', function() {
            updateCountdowns();
            setInterval(updateCountdowns, 1000); // 每秒更新一次
        });
    </script>
</head>
<body>
    <h1>文件浏览器</h1>
    <p>当前路径: %s</p>
    <p>文件过期时间: %d 小时</p>

    <table>
        <thead>
            <tr>
                <th>名称</th>
                <th>大小</th>
                <th>修改时间</th>
                <th>剩余时间</th>
                <th>类型</th>
            </tr>
        </thead>
        <tbody>
`, requestPath, requestPath, expiryHours)

	// 添加返回上级目录链接（如果不是根目录）
	if requestPath != "/" && requestPath != "" {
		parentPath := filepath.Dir(strings.TrimSuffix(requestPath, "/"))
		if parentPath == "." {
			parentPath = "/"
		}
		fmt.Fprintf(w, `
            <tr>
                <td><a href="%s">..</a></td>
                <td>-</td>
                <td>-</td>
                <td>-</td>
                <td>目录</td>
            </tr>
`, parentPath)
	}

	// 显示目录内容
	now := time.Now()
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		size := "-"
		fileType := "目录"
		remainingTime := "-"
		modTime := info.ModTime().Format("2006-01-02 15:04:05")

		if !entry.IsDir() {
			size = fmt.Sprintf("%.2f KB", float64(info.Size())/1024)
			fileType = "文件"

			// 计算文件过期时间
			if expiryHours > 0 {
				expireTime := info.ModTime().Add(time.Duration(expiryHours) * time.Hour)
				expireUnix := expireTime.Unix()

				if expireTime.Before(now) {
					remainingTime = fmt.Sprintf(`<span class="expired">已过期</span>`)
				} else {
					remainingTime = fmt.Sprintf(`<span class="countdown" data-expire-time="%d">计算中...</span>`, expireUnix)
				}
			}
		}

		href := filepath.Join(requestPath, name)
		if entry.IsDir() {
			href += "/"
		}

		fmt.Fprintf(w, `
            <tr>
                <td><a href="%s">%s</a></td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
            </tr>
`, href, name, size, modTime, remainingTime, fileType)
	}

	fmt.Fprintf(w, `
        </tbody>
    </table>

    <div class="upload-section">
        <h2>上传文件</h2>
        <form action="/uploads" method="post" enctype="multipart/form-data">
            <input type="file" name="file" multiple>
            <br>
            <input type="submit" value="上传文件">
        </form>
    </div>
</body>
</html>
`)
}

func handleUpload(w http.ResponseWriter, r *http.Request, uploadDir string) {
	if r.Method != "POST" {
		// 显示上传页面
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>文件上传</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .upload-section { max-width: 500px; margin: 0 auto; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        input[type="file"] { margin: 10px 0; width: 100%%; }
        input[type="submit"] { background-color: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; width: 100%%; }
        input[type="submit"]:hover { background-color: #0056b3; }
        .message { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .success { background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background-color: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
    </style>
</head>
<body>
    <div class="upload-section">
        <h2>文件上传</h2>
        <form action="/uploads" method="post" enctype="multipart/form-data">
            <input type="file" name="file" multiple required>
            <br>
            <input type="submit" value="上传文件">
        </form>
        <br>
        <a href="/">返回文件浏览器</a>
    </div>
</body>
</html>
`)
		return
	}

	// 处理文件上传
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		http.Error(w, "解析表单失败", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		http.Error(w, "没有选择文件", http.StatusBadRequest)
		return
	}

	var uploadedFiles []string
	var failedFiles []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}
		defer file.Close()

		// 创建目标文件
		dst, err := os.Create(filepath.Join(uploadDir, fileHeader.Filename))
		if err != nil {
			file.Close()
			failedFiles = append(failedFiles, fileHeader.Filename)
			continue
		}
		defer dst.Close()

		// 复制文件内容
		_, err = io.Copy(dst, file)
		if err != nil {
			failedFiles = append(failedFiles, fileHeader.Filename)
		} else {
			uploadedFiles = append(uploadedFiles, fileHeader.Filename)
		}
	}

	// 返回上传结果
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>上传结果</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .message { padding: 15px; margin: 10px 0; border-radius: 4px; }
        .success { background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background-color: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .back-link { margin-top: 20px; }
    </style>
</head>
<body>
`)

	if len(uploadedFiles) > 0 {
		fmt.Fprintf(w, `<div class="message success"><strong>成功上传 %d 个文件:</strong><br>%s</div>`,
			len(uploadedFiles), strings.Join(uploadedFiles, "<br>"))
	}

	if len(failedFiles) > 0 {
		fmt.Fprintf(w, `<div class="message error"><strong>上传失败 %d 个文件:</strong><br>%s</div>`,
			len(failedFiles), strings.Join(failedFiles, "<br>"))
	}

	fmt.Fprintf(w, `
    <div class="back-link">
        <a href="/uploads">继续上传</a> |
        <a href="/">返回文件浏览器</a>
    </div>
</body>
</html>
`)
}

// startFileCleanupTask 启动文件清理定时任务
func startFileCleanupTask(workDir string, expiryHours int) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Printf("文件清理任务已启动，每分钟检查一次，文件过期时间: %d 小时", expiryHours)

	for {
		select {
		case <-ticker.C:
			cleanupExpiredFiles(workDir, expiryHours)
		}
	}
}

// cleanupExpiredFiles 清理过期文件
func cleanupExpiredFiles(workDir string, expiryHours int) {
	if expiryHours <= 0 {
		return
	}

	expiryDuration := time.Duration(expiryHours) * time.Hour
	now := time.Now()
	deletedCount := 0

	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理其他文件
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件是否过期
		if now.Sub(info.ModTime()) > expiryDuration {
			if err := os.Remove(path); err != nil {
				log.Printf("删除过期文件失败: %s, 错误: %v", path, err)
			} else {
				log.Printf("已删除过期文件: %s", path)
				deletedCount++
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("文件清理过程中发生错误: %v", err)
	}

	if deletedCount > 0 {
		log.Printf("本次清理完成，共删除 %d 个过期文件", deletedCount)
	}
}
