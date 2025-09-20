#!/bin/bash

# 设置项目名称
PROJECT_NAME="fileserver"

# 创建 releases 目录
mkdir -p releases

# 清理之前的构建文件
echo "清理之前的构建文件..."
rm -rf releases/*

# 定义要编译的平台
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/386"
    "linux/arm"
    "linux/arm64"
    "windows/amd64"
    "windows/386"
    "freebsd/amd64"
    "freebsd/386"
)

echo "开始交叉编译..."

# 使用传统方式逐个编译平台
for platform in "${PLATFORMS[@]}"; do
    echo "编译 $platform ..."

    # 分离 OS 和 ARCH
    IFS='/' read -ra OS_ARCH <<< "$platform"
    OS="${OS_ARCH[0]}"
    ARCH="${OS_ARCH[1]}"

    # 确定可执行文件扩展名
    EXT=""
    if [ "$OS" = "windows" ]; then
        EXT=".exe"
    fi

    # 创建输出目录
    OUTPUT_DIR="releases/${OS}_${ARCH}"
    mkdir -p "$OUTPUT_DIR"

    # 编译
    GOOS=$OS GOARCH=$ARCH go build -o "$OUTPUT_DIR/$PROJECT_NAME$EXT" .

    if [ $? -eq 0 ]; then
        echo "  ✓ 编译成功: $OUTPUT_DIR/$PROJECT_NAME$EXT"
    else
        echo "  ✗ 编译失败: $platform"
    fi
done

# 复制配置文件到各个平台目录
echo "复制配置文件到各个平台目录..."

# 遍历所有生成的目录
for dir in releases/*/; do
    if [ -d "$dir" ]; then
        platform=$(basename "$dir")
        echo "处理目录: $platform"

        # 复制配置文件
        if [ -f "config.json" ]; then
            cp config.json "$dir"
            echo "  - 已复制 config.json"
        else
            # 创建默认配置文件
            cat > "$dir/config.json" << EOF
{
    "port": "9527",
    "workdir": "./files",
    "uploaddir": "./uploads"
}
EOF
            echo "  - 已创建默认 config.json"
        fi

        # 创建 README 文件
        binary_name="$PROJECT_NAME"
        if [[ $platform == windows* ]]; then
            binary_name="$PROJECT_NAME.exe"
        fi

        cat > "$dir/README.txt" << EOF
文件服务器 - $platform

使用方法：
1. Linux/macOS/FreeBSD: ./$binary_name -port 9527 -workdir ./uploads -uploaddir ./uploads
2. Windows: $binary_name -port 9527 -workdir ./uploads -uploaddir ./uploads

或者使用配置文件：
./$binary_name -config config.json

默认配置：
- 端口: 9527
- 工作目录: ./uploads
- 上传目录: ./uploads

请确保上传目录存在，或者程序会自动创建。

启动后访问：
- 文件浏览: http://localhost:9527/
- 文件上传: http://localhost:9527/uploads
EOF
        echo "  - 已创建 README.txt"
    fi
done

# 显示构建结果
echo ""
echo "构建完成！生成的文件位于 releases 目录下："
echo "=========================================="
for dir in releases/*/; do
    if [ -d "$dir" ]; then
        platform=$(basename "$dir")
        # 查找可执行文件
        if [[ $platform == windows* ]]; then
            binary=$(ls "$dir" | grep "\.exe$" | head -1)
        else
            binary=$(ls "$dir" | grep "^fileserver$" | head -1)
        fi
        if [ -n "$binary" ]; then
            echo "$platform: $binary"
        fi
    fi
done

echo ""
echo "总构建目录数量: $(ls -d releases/*/ 2>/dev/null | wc -l | tr -d ' ')"
