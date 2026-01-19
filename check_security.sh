#!/bin/bash

# 安全检查脚本 - 检查是否有敏感信息被意外提交

echo "🔒 执行安全检查..."
echo "===================="

# 检查是否有敏感文件被git跟踪
echo "📁 检查敏感文件..."
SENSITIVE_FILES=$(git ls-files | grep -E "(fofa_config\.json|fofa.*result.*\.json|secrets\.json|\.db$|\.log$)" | grep -v "\.example$")

if [ -n "$SENSITIVE_FILES" ]; then
    echo "❌ 发现被git跟踪的敏感文件:"
    echo "$SENSITIVE_FILES"
    echo ""
    echo "请使用以下命令移除:"
    echo "git rm --cached <filename>"
    exit 1
else
    echo "✅ 没有发现被跟踪的敏感文件"
fi

# 检查是否有真实邮箱地址在配置文件中
echo ""
echo "📧 检查真实邮箱地址..."
REAL_EMAILS=$(git ls-files -z | xargs -0 grep -l "@" | grep "\.json$" | xargs grep "@" | grep -v "example\.com" | grep -v "your_email")

if [ -n "$REAL_EMAILS" ]; then
    echo "⚠️  发现可能的真实邮箱地址:"
    echo "$REAL_EMAILS"
    echo ""
    echo "请检查这些文件是否包含敏感信息"
else
    echo "✅ 没有发现真实邮箱地址"
fi

# 检查是否有可疑的API密钥格式
echo ""
echo "🔑 检查API密钥格式..."
SUSPICIOUS_KEYS=$(git ls-files -z | xargs -0 grep -l "[a-f0-9]\{32\}" | grep "\.json$" | xargs grep "[a-f0-9]\{32\}" | grep -v "your_fofa_api_key")

if [ -n "$SUSPICIOUS_KEYS" ]; then
    echo "⚠️  发现可疑的API密钥格式:"
    echo "$SUSPICIOUS_KEYS"
    echo ""
    echo "请检查这些是否为真实的API密钥"
else
    echo "✅ 没有发现可疑的API密钥"
fi

# 检查.gitignore文件是否存在
echo ""
echo "📋 检查.gitignore文件..."
if [ -f ".gitignore" ]; then
    echo "✅ .gitignore文件存在"
    
    # 检查关键忽略规则
    if grep -q "fofa_config.json" .gitignore; then
        echo "✅ FOFA配置文件已被忽略"
    else
        echo "❌ FOFA配置文件未被忽略"
    fi
    
    if grep -q "*.log" .gitignore; then
        echo "✅ 日志文件已被忽略"
    else
        echo "❌ 日志文件未被忽略"
    fi
    
    if grep -q "*.db" .gitignore; then
        echo "✅ 数据库文件已被忽略"
    else
        echo "❌ 数据库文件未被忽略"
    fi
else
    echo "❌ .gitignore文件不存在"
    exit 1
fi

echo ""
echo "🎉 安全检查完成！"
echo ""
echo "💡 提示:"
echo "- 提交前请确保没有包含真实的API密钥"
echo "- 使用示例配置文件格式"
echo "- 定期轮换API密钥"