#!/bin/bash

# 项目打包脚本
# 用于将整个项目打包成tar.gz文件，以便在CentOS上部署

echo "=== 项目打包脚本 ==="
echo "执行时间：$(date)"
echo ""

# 项目名称和版本
PROJECT_NAME="industrial-protocol-services"
VERSION="1.0.0"
ARCHIVE_NAME="${PROJECT_NAME}-${VERSION}.tar.gz"

# 临时目录结构
TEMP_DIR="../temp-package"
FINAL_DIR="${TEMP_DIR}/${PROJECT_NAME}-${VERSION}"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "1. 清理旧的打包文件..."
rm -f "../${ARCHIVE_NAME}" 2>/dev/null
rm -rf "${TEMP_DIR}" 2>/dev/null
echo "   ✓ 旧文件已清理"
echo ""

echo "2. 创建临时目录结构..."
mkdir -p "${FINAL_DIR}" "${FINAL_DIR}/templates" "${FINAL_DIR}/logs"
echo "   ✓ 临时目录已创建：${FINAL_DIR}"
echo ""

echo "3. 复制项目文件到临时目录..."
# 手动选择需要复制的文件
# 核心Python文件
python_files=("cve_2021_4161_poc.py" "diagnose_redis.py" "kafka_client.py" "kafka_server.py" "modbus_client.py" "modbus_server.py" "mysql_client.py" "mysql_server.py" "redis_client.py" "redis_server.py" "service_config.json" "service_manager.py" "system_logger.py" "vulnerable_modbus_gateway.py" "web_service_manager.py")

# 启动脚本
start_scripts=("start_services_centos.sh" "start_services_mac.sh")

# 模板文件
template_files=("templates/index.html")

# 复制核心Python文件
for file in "${python_files[@]}"; do
    if [ -f "${SCRIPT_DIR}/${file}" ]; then
        cp "${SCRIPT_DIR}/${file}" "${FINAL_DIR}/"
    fi
done

# 复制启动脚本
for file in "${start_scripts[@]}"; do
    if [ -f "${SCRIPT_DIR}/${file}" ]; then
        cp "${SCRIPT_DIR}/${file}" "${FINAL_DIR}/"
        chmod +x "${FINAL_DIR}/${file}"
    fi
done

# 复制模板文件
for file in "${template_files[@]}"; do
    src="${SCRIPT_DIR}/${file}"
    dst="${FINAL_DIR}/${file}"
    if [ -f "${src}" ]; then
        cp "${src}" "${dst}"
    fi
done

echo "   ✓ 项目文件已复制"
echo ""

echo "4. 打包项目..."
cd "${TEMP_DIR}"
tar -czvf "../${ARCHIVE_NAME}" "${PROJECT_NAME}-${VERSION}" || {
    echo "   ✗ 打包失败"
    rm -rf "${TEMP_DIR}"
    exit 1
}
echo "   ✓ 项目已打包为：${ARCHIVE_NAME}"
echo ""

echo "5. 清理临时目录..."
rm -rf "${TEMP_DIR}"
echo "   ✓ 临时目录已清理"
echo ""

echo "=== 打包完成 ==="
echo "打包文件：${ARCHIVE_NAME}"
echo "文件大小：$(du -h "../${ARCHIVE_NAME}" | awk '{print $1}')"
echo ""
echo "使用以下命令在CentOS上部署："
echo "1. 将${ARCHIVE_NAME}上传到CentOS服务器"
echo "2. 在CentOS服务器上执行："
echo "   tar -xzvf ${ARCHIVE_NAME}"
echo "   cd ${PROJECT_NAME}-${VERSION}"
echo "   ./start_services_centos.sh"
echo ""
echo "部署前建议检查："
echo "- 确保CentOS服务器已安装Python 3"
echo "- 确保CentOS服务器已安装curl命令"
echo "- 确保端口9999未被占用"
echo ""
echo "文件列表："
cd "${SCRIPT_DIR}/.."
tar -tzf "${ARCHIVE_NAME}" | sort
echo ""
