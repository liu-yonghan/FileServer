# 文件服务器

一个简单易用的 Go 语言文件服务器，支持文件浏览、上传和自动清理过期文件。

## 功能特性

- 📁 **文件浏览**: 通过 Web 界面浏览服务器文件
- 📤 **文件上传**: 支持单文件和多文件上传
- ⏰ **倒计时显示**: 实时显示文件过期倒计时，直观了解文件状态
- 🗑️ **自动清理**: 定时清理过期文件，节省存储空间
- ⚙️ **灵活配置**: 支持配置文件和命令行参数
- 🔒 **安全防护**: 防止路径遍历攻击
- 🌐 **跨平台**: 支持 Windows、Linux、macOS 等多种平台

## 快速开始

### 1. 下载和安装

从 [Releases](releases/) 页面下载适合你系统的预编译版本，或者从源码编译：

```bash
git clone <repository-url>
cd fileserver
go build -o fileserver main.go
```

### 2. 运行服务器

使用默认配置运行：
```bash
./fileserver
```

使用自定义配置：
```bash
./fileserver -port 9527 -workdir ./files -expiry 24
```

### 3. 访问服务

- 文件浏览: http://localhost:8080/
- 文件上传: http://localhost:8080/uploads

## 配置说明

### 配置文件 (config.json)

```json
{
  "port": "9527",
  "workdir": "./uploads",
  "uploaddir": "./uploads",
  "file_expiry_hours": 2
}
```

### 配置参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `port` | 服务器端口 | 8080 |
| `workdir` | 文件浏览的工作目录 | ./uploads |
| `uploaddir` | 文件上传的目标目录 | ./uploads |
| `file_expiry_hours` | 文件过期时间（小时） | 2 |

### 命令行参数

```bash
./fileserver [选项]

选项:
  -port string        服务器端口 (默认 "8080")
  -workdir string     工作目录 (默认 "./uploads")
  -uploaddir string   上传目录 (默认 "./uploads")
  -expiry int         文件过期时间（小时） (默认 2)
  -config string      配置文件路径 (默认 "./config.json")
```

## 文件过期管理

### 倒计时显示
- **实时更新**: 文件列表中显示每个文件的剩余时间，每秒自动更新
- **状态标识**:
  - 正常文件：显示剩余时间（如：2时30分15秒）
  - 即将过期：小于1小时时显示橙色警告
  - 已过期：显示红色"已过期"标识
- **直观界面**: 一目了然地了解所有文件的过期状态

### 自动清理
服务器启动后会自动启动文件清理任务：

- **检查频率**: 每分钟检查一次
- **清理规则**: 删除修改时间超过设定过期时间的文件
- **日志记录**: 清理过程会记录到控制台日志
- **安全性**: 只删除文件，不删除目录

### 清理日志示例

```
2024/01/20 10:30:00 文件清理任务已启动，每分钟检查一次，文件过期时间: 2 小时
2024/01/20 10:31:00 已删除过期文件: /uploads/old_file.txt
2024/01/20 10:31:00 本次清理完成，共删除 1 个过期文件
```

## 使用示例

### 1. 基本使用

```bash
# 启动服务器
./fileserver

# 在浏览器中访问
open http://localhost:8080
```

### 2. 自定义配置

```bash
# 使用自定义端口和目录
./fileserver -port 9527 -workdir ./files -uploaddir ./uploads

# 设置文件 24 小时后过期
./fileserver -expiry 24
```

### 3. 使用配置文件

创建 `config.json`:
```json
{
  "port": "9527",
  "workdir": "./data",
  "uploaddir": "./data/uploads",
  "file_expiry_hours": 48
}
```

运行：
```bash
./fileserver -config ./config.json
```

## 安全注意事项

1. **网络访问**: 默认只监听本地地址，如需外网访问请谨慎配置防火墙
2. **文件权限**: 确保运行用户对工作目录有读写权限
3. **存储空间**: 定期检查磁盘空间，合理设置文件过期时间
4. **上传限制**: 当前单文件上传限制为 32MB

## 构建说明

### 本地构建

```bash
go build -o fileserver main.go
```

### 交叉编译

```bash
# Windows 64位
GOOS=windows GOARCH=amd64 go build -o fileserver.exe main.go

# Linux 64位
GOOS=linux GOARCH=amd64 go build -o fileserver main.go

# macOS 64位
GOOS=darwin GOARCH=amd64 go build -o fileserver main.go
```

## 故障排除

### 常见问题

1. **端口被占用**
   ```
   解决方案: 使用 -port 参数指定其他端口
   ```

2. **权限不足**
   ```
   解决方案: 确保对工作目录有读写权限
   ```

3. **文件上传失败**
   ```
   解决方案: 检查上传目录是否存在且可写
   ```

### 日志查看

服务器运行时会在控制台输出详细日志，包括：
- 服务器启动信息
- 文件清理日志
- 错误信息

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.2.0
- ⏰ 新增文件过期倒计时显示功能
- 🎨 优化界面，增加过期状态颜色标识
- ⚡ 实时更新倒计时，提升用户体验

### v1.1.0
- ✨ 新增文件自动过期清理功能
- ⚙️ 支持配置文件过期时间
- 📝 完善文档和使用说明

### v1.0.0
- 🎉 初始版本发布
- 📁 支持文件浏览和上传
- 🔧 支持配置文件和命令行参数
