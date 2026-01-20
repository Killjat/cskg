# 非标准端口服务探测演示

## 问题场景

在实际网络环境中，管理员经常将服务部署在非标准端口上，原因包括：
- **安全考虑**: 避免自动化扫描和攻击
- **端口冲突**: 标准端口被其他服务占用
- **负载均衡**: 在不同端口运行多个服务实例
- **隐蔽性**: 降低服务被发现的概率

## 传统方法的局限性

传统的端口扫描工具通常基于端口号来推断服务类型：
- 22端口 → SSH服务
- 80端口 → HTTP服务
- 443端口 → HTTPS服务
- 3306端口 → MySQL服务

但这种方法在面对非标准端口部署时会失效。

## 全面探测的优势

我们的网络探测引擎通过发送所有协议的探测包，能够准确识别任意端口上的真实服务。

### 示例1: 22端口上的HTTP服务

```bash
# 传统方法会认为这是SSH服务
nmap -p 22 target.com

# 我们的全面探测
./network_probe -target target.com:22 -probe-mode all
```

**可能的结果**:
```
🎯 目标: target.com:22
✅ 成功探测: 3/14

1. ❌ SSHVersionExchange (ssh) - connection refused
2. ✅ GetRequest (http) - 耗时: 45ms
   📄 Banner: "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0..."
   🏷️  产品: nginx v1.18.0 (置信度: 90%)
3. ✅ HTTPOptions (http) - 耗时: 42ms
```

### 示例2: 8080端口上的数据库服务

```bash
# 全面探测8080端口
./network_probe -target db-server:8080 -probe-mode all -verbose
```

**可能发现**:
```
✅ MySQLGreeting (mysql) - 耗时: 23ms
📄 Banner: "5.7.35-log\x00..."
🏷️  产品: MySQL v5.7.35 (置信度: 90%)
```

### 示例3: 443端口上的SSH服务

```bash
# 探测HTTPS端口，但发现SSH服务
./network_probe -target server:443 -probe-mode all
```

**可能结果**:
```
✅ SSHVersionExchange (ssh) - 耗时: 15ms
📄 Banner: "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5"
🏷️  产品: OpenSSH v8.2p1 (置信度: 95%)
```

## 探测模式对比

### Port模式 (快速但可能遗漏)
```bash
./network_probe -target mystery-server:2222 -probe-mode port
# 只会尝试与2222端口相关的探测（可能没有）
```

### All模式 (全面但较慢)
```bash
./network_probe -target mystery-server:2222 -probe-mode all
# 尝试所有14种协议探测，确保不遗漏任何服务
```

### Smart模式 (平衡)
```bash
./network_probe -target mystery-server:2222 -probe-mode smart
# 优先尝试常见协议，平衡速度和覆盖面
```

## 实际应用场景

### 1. 安全评估
```bash
# 发现隐藏在非标准端口的服务
./network_probe -target 192.168.1.100:8888 -probe-mode all
```

### 2. 网络资产清点
```bash
# 识别服务器上所有端口的真实服务
for port in 22 80 443 8080 8443 9000; do
    ./network_probe -target server:$port -probe-mode all
done
```

### 3. 故障排查
```bash
# 确认服务是否在预期端口运行
./network_probe -target app-server:3000 -probe-mode all -verbose
```

## 最佳实践

1. **使用all模式进行初始探测**: 确保不遗漏任何服务
2. **结合多种探测模式**: 根据场景选择合适的模式
3. **关注响应内容**: 即使探测"失败"，响应内容也可能透露服务信息
4. **并发控制**: 调整并发数避免触发防护机制
5. **超时设置**: 根据网络环境调整超时时间

## 防护建议

对于系统管理员，了解这种探测方法有助于：
- **评估服务暴露风险**: 即使在非标准端口也可能被发现
- **配置防火墙规则**: 基于协议而非仅端口进行过滤
- **监控异常连接**: 检测针对服务的协议探测行为
- **服务加固**: 对非标准端口服务进行额外保护