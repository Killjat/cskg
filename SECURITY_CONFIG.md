# 安全配置说明

## 🔒 敏感信息保护

本项目包含多个需要API密钥和凭据的组件，为了保护敏感信息，我们采用了以下安全措施：

### 📁 被保护的文件

以下文件包含敏感信息，已被 `.gitignore` 忽略，不会上传到git仓库：

```
# FOFA 配置文件
**/fofa_config.json
**/fofa-config.json  
**/fofa_credentials.json

# FOFA 测试结果（可能包含敏感目标信息）
**/fofa_test_result_*.json
**/fofa_api_test_result_*.json

# 数据库文件
**/*.db
**/*.sqlite

# 日志文件
**/*.log
**/logs/*.log

# 其他敏感配置
**/secrets.json
**/private_config.json
```

### 🛠️ 配置方法

#### 1. FOFA API 配置

**fofa-api-test 项目**:
```bash
cd fofa-api-test
cp fofa_config.json.example fofa_config.json
# 编辑 fofa_config.json，填入真实的FOFA凭据
```

**network_probe 项目**:
```bash
cd network_probe  
cp fofa_config.json.example fofa_config.json
# 编辑 fofa_config.json，填入真实的FOFA凭据
```

#### 2. 配置文件格式

```json
{
  "email": "your_fofa_email@example.com",
  "key": "your_fofa_api_key_here",
  "base_url": "https://fofa.info/api/v1/search/all"
}
```

### 🔑 获取FOFA凭据

1. 访问 [FOFA官网](https://fofa.info)
2. 注册并登录账户
3. 进入个人中心 → API管理
4. 获取你的邮箱和API Key
5. 将凭据填入配置文件

### ⚠️ 安全注意事项

1. **永远不要**将包含真实凭据的配置文件提交到git
2. **定期轮换**API密钥，特别是在怀疑泄露时
3. **限制权限**，只给予必要的API访问权限
4. **监控使用**，定期检查API使用情况
5. **团队协作**时，通过安全渠道分享凭据

### 🚨 如果凭据泄露

如果不小心将包含真实凭据的文件提交到了git：

1. **立即轮换**所有相关的API密钥
2. **清理git历史**：
   ```bash
   # 从git历史中完全删除敏感文件
   git filter-branch --force --index-filter \
     'git rm --cached --ignore-unmatch path/to/sensitive/file' \
     --prune-empty --tag-name-filter cat -- --all
   
   # 强制推送（谨慎操作）
   git push origin --force --all
   ```
3. **通知团队**成员更新本地仓库
4. **检查日志**，确认是否有异常API使用

### 📋 检查清单

在提交代码前，请确认：

- [ ] 没有包含真实的API密钥
- [ ] 没有包含真实的邮箱地址  
- [ ] 没有包含敏感的测试结果
- [ ] 所有配置文件都使用示例格式
- [ ] `.gitignore` 文件已正确配置

### 🔧 自动化检查

你可以使用以下命令检查是否有敏感信息：

```bash
# 检查是否有真实邮箱地址
grep -r "@" --include="*.json" . | grep -v "example.com"

# 检查是否有可疑的API密钥格式
grep -r "[a-f0-9]{32}" --include="*.json" .

# 检查git状态
git status | grep -E "(fofa|config|secret)"
```

### 📞 联系方式

如果发现安全问题或需要帮助，请联系项目维护者。

---

**记住：安全是每个人的责任！** 🛡️