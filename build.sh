#!/bin/bash

# 设置版本号
VERSION="1.0.0"

# 设置输出目录
OUTPUT_DIR="build"

# 创建输出目录
mkdir -p ${OUTPUT_DIR}

# 编译函数
build() {
    local GOOS=$1
    local GOARCH=$2
    local SUFFIX=$3
    
    echo "Building for ${GOOS}/${GOARCH}..."
    
    # 设置输出文件名
    local OUTPUT_NAME="device-upgrade-${VERSION}-${GOOS}-${GOARCH}${SUFFIX}"
    local OUTPUT_PATH="${OUTPUT_DIR}/${OUTPUT_NAME}"
    
    # 如果是 macOS，创建 .app 包
    if [ "$GOOS" = "darwin" ]; then
        echo "Creating macOS app bundle for ${GOARCH}..."
        
        # 设置编译标志
        CGO_ENABLED=1 GOOS=${GOOS} GOARCH=${GOARCH} \
        go build -tags no_native_menu \
            -ldflags="-s -w" \
            -o DeviceUpgrade ./cmd/main.go
        
        if [ $? -eq 0 ]; then
            # 创建应用包结构
            APP_NAME="DeviceUpgrade-${GOARCH}.app"
            APP_PATH="${OUTPUT_DIR}/${APP_NAME}"
            mkdir -p "${APP_PATH}/Contents/MacOS"
            mkdir -p "${APP_PATH}/Contents/Resources"
            
            # 移动可执行文件
            mv DeviceUpgrade "${APP_PATH}/Contents/MacOS/"
            
            # 复制配置文件到 Resources 目录
            cp config/config.yaml "${APP_PATH}/Contents/Resources/config.yaml"
            
            # 创建 Info.plist
            cat > "${APP_PATH}/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>DeviceUpgrade</string>
    <key>CFBundleIdentifier</key>
    <string>net.yingzhu.deviceupgrade</string>
    <key>CFBundleName</key>
    <string>DeviceUpgrade</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
</dict>
</plist>
EOF
            
            # 设置权限
            chmod +x "${APP_PATH}/Contents/MacOS/DeviceUpgrade"
            
            # 对应用进行代码签名
            if command -v codesign &> /dev/null; then
                echo "Signing application..."
                codesign --force --deep --sign - "${APP_PATH}"
            fi
            
            echo "Successfully created ${APP_PATH}"
        else
            echo "Failed to build for macOS ${GOARCH}"
        fi
        
    else
        # Windows 平台编译
        echo "Building Windows ${GOARCH} version..."
        
        # 设置CGO_ENABLED=1以支持GUI
        CGO_ENABLED=1 GOOS=${GOOS} GOARCH=${GOARCH} CC=x86_64-w64-mingw32-gcc \
        go build -tags no_native_menu \
            -ldflags="-s -w -H=windowsgui" \
            -o ${OUTPUT_PATH} ./cmd/main.go
        
        if [ $? -eq 0 ]; then
            echo "Successfully built ${OUTPUT_NAME}"
            
            # 创建发布目录
            RELEASE_DIR="${OUTPUT_DIR}/device-upgrade-${VERSION}-windows-${GOARCH}"
            mkdir -p "${RELEASE_DIR}"
            
            # 复制文件到发布目录
            cp "${OUTPUT_PATH}" "${RELEASE_DIR}/DeviceUpgrade.exe"
            cp config/config.yaml "${RELEASE_DIR}/"
            
            # 创建压缩包
            if command -v zip &> /dev/null; then
                # 使用 zip（在 Unix 系统上）
                (cd "${OUTPUT_DIR}" && zip -r "device-upgrade-${VERSION}-windows-${GOARCH}.zip" "device-upgrade-${VERSION}-windows-${GOARCH}")
            else
                # 使用 7z（如果可用）
                if command -v 7z &> /dev/null; then
                    7z a "${OUTPUT_DIR}/device-upgrade-${VERSION}-windows-${GOARCH}.zip" "${RELEASE_DIR}/*"
                else
                    echo "Warning: Neither zip nor 7z found. Skipping archive creation."
                fi
            fi
            
            # 清理临时文件
            rm -f "${OUTPUT_PATH}"
            rm -rf "${RELEASE_DIR}"
        else
            echo "Failed to build for Windows ${GOARCH}"
        fi
    fi
}

# 清理旧的构建文件
rm -rf ${OUTPUT_DIR}/*

# 编译 macOS 版本
build "darwin" "amd64" ""        # Mac Intel
build "darwin" "arm64" ""        # Mac M1

# 编译 Windows 版本 (只编译最常用的版本)
build "windows" "amd64" ".exe"   # Windows 64位

echo "Build complete! Check the ${OUTPUT_DIR} directory for the applications."

# 显示构建结果
echo -e "\nBuild results:"
ls -lh ${OUTPUT_DIR}