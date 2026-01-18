# 远程服务不可用故障排查指南

## 检查项目状态

### 本地测试结果
- ✅ 服务已在8080端口运行（PID 14875）
- ✅ 本地可以正常访问：`curl http://localhost:8080` 返回了HTML内容
- ✅ 服务绑定到所有网络接口（0.0.0.0:8080）
- ✅ 已配置CORS支持

## 远程部署故障排查步骤

### 步骤1：检查服务器基本状态

```bash
# 登录服务器
ssh root@121.43.143.169

# 检查服务器负载
uptime

# 检查内存使用
free -h

# 检查磁盘空间
df -h
```

### 步骤2：检查服务状态

```bash
# 查看服务运行状态
systemctl status web-clone-tool.service

# 查看服务日志（获取详细错误信息）
journalctl -u web-clone-tool.service -n 50

# 查看实时日志
journalctl -u web-clone-tool.service -f
```

### 步骤3：检查端口占用

```bash
# 检查8080端口是否被占用
lsof -i :8080

# 如果lsof不可用，尝试使用netstat
netstat -tuln | grep 8080

# 或者使用ss命令
ss -tuln | grep 8080
```

### 步骤4：检查防火墙配置

```bash
# 检查防火墙状态
firewall-cmd --state

# 查看开放的端口
firewall-cmd --list-ports

# 如果8080端口未开放，添加开放规则
firewall-cmd --add-port=8080/tcp --permanent
firewall-cmd --reload

# 检查是否生效
firewall-cmd --list-ports
```

### 步骤5：检查SELinux配置

```bash
# 查看SELinux状态
grep -i selinux /etc/selinux/config

# 临时禁用SELinux（用于测试）
setenforce 0

# 如果SELinux是Enforcing，检查是否有相关的审计日志
audit2why < /var/log/audit/audit.log

# 允许http服务访问网络
setsebool -P httpd_can_network_connect 1
```

### 步骤6：检查Python依赖

```bash
# 进入应用目录
cd /opt/web-clone-tool

# 检查Python版本
python3 --version

# 检查依赖是否安装成功
pip3 list | grep -E "fastapi|uvicorn|beautifulsoup4|requests"

# 重新安装依赖（如果有问题）
pip3 install --upgrade pip
pip3 install -r requirements.txt
```

### 步骤7：手动启动服务测试

```bash
# 先停止systemd服务
systemctl stop web-clone-tool.service

# 手动启动服务，查看输出
python3 app.py

# 测试本地访问
curl http://localhost:8080
```

### 步骤8：检查网络连接

```bash
# 从服务器内部测试
curl -I http://localhost:8080
curl -I http://127.0.0.1:8080
curl -I http://$(hostname -I | cut -d' ' -f1):8080

# 测试从外部访问（在本地机器上执行）
curl -v http://121.43.143.169:8080
ping 121.43.143.169
```

### 步骤9：检查服务绑定地址

```bash
# 查看服务是否绑定到所有网络接口
netstat -tuln | grep 8080
# 应该显示 0.0.0.0:8080 或 :::8080
```

## 常见错误及解决方案

### 错误1：服务无法启动

**症状**：`systemctl status web-clone-tool.service` 显示服务失败

**解决方案**：
1. 检查服务日志：`journalctl -u web-clone-tool.service -n 50`
2. 检查Python依赖是否正确安装
3. 检查app.py文件是否有语法错误：`python3 -m py_compile app.py`

### 错误2：端口被占用

**症状**：`lsof -i :8080` 显示多个进程占用8080端口

**解决方案**：
1. 找到占用端口的进程：`lsof -i :8080`
2. 终止占用端口的进程：`kill -9 <PID>`
3. 重启服务：`systemctl restart web-clone-tool.service`

### 错误3：防火墙阻止访问

**症状**：本地可以访问，但外部无法访问

**解决方案**：
1. 开放8080端口：`firewall-cmd --add-port=8080/tcp --permanent && firewall-cmd --reload`
2. 检查云服务器安全组是否开放了8080端口

### 错误4：Python版本不兼容

**症状**：服务启动时出现语法错误或模块导入错误

**解决方案**：
1. 确保使用Python 3.10+：`python3 --version`
2. 检查requirements.txt中的依赖版本
3. 重新安装依赖：`pip3 install -r requirements.txt`

### 错误5：SELinux限制

**症状**：服务运行但无法访问

**解决方案**：
1. 临时禁用SELinux：`setenforce 0`（测试）
2. 允许http服务访问网络：`setsebool -P httpd_can_network_connect 1`
3. 查看审计日志：`audit2why < /var/log/audit/audit.log`

## 重新部署步骤

如果以上排查步骤无法解决问题，可以尝试重新部署：

```bash
# 登录服务器
ssh root@121.43.143.169

# 停止并禁用现有服务
systemctl stop web-clone-tool.service
systemctl disable web-clone-tool.service

# 删除现有应用目录
rm -rf /opt/web-clone-tool
rm -f /etc/systemd/system/web-clone-tool.service

# 重新上传最新压缩包
scp web-clone-tool-20260112_214009.tar.gz root@121.43.143.169:/root/

# 解压并部署
tar -zxvf web-clone-tool-20260112_214009.tar.gz
cd web-clone-tool
chmod +x deploy_centos.sh
sudo ./deploy_centos.sh
```

## 验证部署成功

```bash
# 检查服务状态
systemctl status web-clone-tool.service

# 检查端口监听
lsof -i :8080

# 测试本地访问
curl -I http://localhost:8080

# 测试外部访问（在本地机器上执行）
curl -I http://121.43.143.169:8080
```

## 联系支持

如果以上步骤都无法解决问题，请收集以下信息并提供：
1. 服务器的操作系统版本：`cat /etc/os-release`
2. Python版本：`python3 --version`
3. 服务日志：`journalctl -u web-clone-tool.service -n 100`
4. 防火墙状态：`firewall-cmd --list-all`
5. SELinux状态：`sestatus`
6. 端口状态：`netstat -tuln | grep 8080`

通过以上详细的排查步骤，应该可以定位并解决远程服务不可用的问题。