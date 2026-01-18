const express = require('express');
const fs = require('fs');
const path = require('path');
const cloneWebPage = require('./clone');

const app = express();
const PORT = 3000;

// 静态文件服务
app.use(express.static(path.join(__dirname, 'public')));
app.use('/cloned', express.static(path.join(__dirname, 'cloned_pages')));

// 解析JSON请求体
app.use(express.json());

// 主页 - 提供克隆功能界面
app.get('/', (req, res) => {
  res.send(`
    <!DOCTYPE html>
    <html lang="zh-CN">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>网页克隆工具</title>
      <style>
        body {
          font-family: Arial, sans-serif;
          max-width: 800px;
          margin: 0 auto;
          padding: 20px;
          background-color: #f5f5f5;
        }
        h1 {
          color: #333;
          text-align: center;
        }
        .container {
          background-color: white;
          padding: 20px;
          border-radius: 8px;
          box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
        }
        .form-group {
          margin-bottom: 20px;
        }
        label {
          display: block;
          margin-bottom: 8px;
          font-weight: bold;
        }
        input[type="url"] {
          width: 100%;
          padding: 10px;
          font-size: 16px;
          border: 1px solid #ddd;
          border-radius: 4px;
        }
        button {
          background-color: #4CAF50;
          color: white;
          padding: 10px 20px;
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 16px;
        }
        button:hover {
          background-color: #45a049;
        }
        .result {
          margin-top: 20px;
          padding: 15px;
          border-radius: 4px;
        }
        .success {
          background-color: #d4edda;
          color: #155724;
          border: 1px solid #c3e6cb;
        }
        .error {
          background-color: #f8d7da;
          color: #721c24;
          border: 1px solid #f5c6cb;
        }
        .cloned-list {
          margin-top: 30px;
        }
        .cloned-item {
          margin-bottom: 15px;
          padding: 15px;
          background-color: #e9ecef;
          border-radius: 4px;
        }
        .cloned-item h3 {
          margin: 0 0 10px 0;
        }
        .cloned-item .links {
          margin-top: 10px;
        }
        .cloned-item a {
          margin-right: 15px;
          color: #007bff;
          text-decoration: none;
        }
        .cloned-item a:hover {
          text-decoration: underline;
        }
      </style>
    </head>
    <body>
      <div class="container">
        <h1>网页克隆工具</h1>
        <div class="form-group">
          <label for="url">输入要克隆的URL：</label>
          <input type="url" id="url" placeholder="https://example.com" required>
        </div>
        <button onclick="clonePage()">克隆网页</button>
        <div id="result" class="result" style="display: none;"></div>
        
        <div class="cloned-list">
          <h2>已克隆的页面</h2>
          <div id="clonedPages"></div>
        </div>
      </div>
      
      <script>
        // 克隆页面功能
        async function clonePage() {
          const url = document.getElementById('url').value;
          const resultDiv = document.getElementById('result');
          
          if (!url) {
            resultDiv.className = 'result error';
            resultDiv.innerHTML = '请输入有效的URL';
            resultDiv.style.display = 'block';
            return;
          }
          
          resultDiv.className = 'result success';
          resultDiv.innerHTML = '正在克隆...';
          resultDiv.style.display = 'block';
          
          try {
            const response = await fetch('/api/clone', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json'
              },
              body: JSON.stringify({ url })
            });
            
            const data = await response.json();
            
            if (data.success) {
              resultDiv.className = 'result success';
              resultDiv.innerHTML = `
                <h3>克隆成功！</h3>
                <p>标题：${data.title}</p>
                <p>保存目录：${data.saveDir}</p>
                <div class="links">
                  <a href="${data.files.simple}" target="_blank">查看简化版</a>
                  <a href="${data.files.full}" target="_blank">查看完整版</a>
                  <a href="${data.files.info}" target="_blank">查看提取信息</a>
                </div>
              `;
            } else {
              resultDiv.className = 'result error';
              resultDiv.innerHTML = `克隆失败：${data.error}`;
            }
          } catch (error) {
            resultDiv.className = 'result error';
            resultDiv.innerHTML = `克隆失败：${error.message}`;
          }
          
          // 刷新已克隆页面列表
          loadClonedPages();
        }
        
        // 加载已克隆页面列表
        async function loadClonedPages() {
          const response = await fetch('/api/cloned-pages');
          const pages = await response.json();
          const container = document.getElementById('clonedPages');
          
          if (pages.length === 0) {
            container.innerHTML = '<p>暂无克隆页面</p>';
            return;
          }
          
          container.innerHTML = pages.map(page => `
            <div class="cloned-item">
              <h3>${page.title}</h3>
              <p>URL: <a href="${page.url}" target="_blank">${page.url}</a></p>
              <p>克隆时间: ${new Date(page.timestamp).toLocaleString()}</p>
              <div class="links">
                <a href="${page.files.simple}" target="_blank">查看简化版</a>
                <a href="${page.files.full}" target="_blank">查看完整版</a>
                <a href="${page.files.info}" target="_blank">查看提取信息</a>
              </div>
            </div>
          `).join('');
        }
        
        // 页面加载时初始化
        window.onload = loadClonedPages;
      </script>
    </body>
    </html>
  `);
});

// API - 克隆网页
app.post('/api/clone', async (req, res) => {
  const { url } = req.body;
  if (!url) {
    return res.status(400).json({ success: false, error: '请提供URL' });
  }
  
  const result = await cloneWebPage(url);
  if (result.success) {
    // 转换为可访问的URL路径
    const relativePath = path.relative(__dirname, result.saveDir);
    result.files = {
      simple: `/cloned/${relativePath}/simple.html`,
      full: `/cloned/${relativePath}/full.html`,
      info: `/cloned/${relativePath}/info.json`
    };
  }
  res.json(result);
});

// API - 获取已克隆页面列表
app.get('/api/cloned-pages', (req, res) => {
  const clonedDir = path.join(__dirname, 'cloned_pages');
  const pages = [];
  
  try {
    if (fs.existsSync(clonedDir)) {
      const dirs = fs.readdirSync(clonedDir, { withFileTypes: true })
        .filter(dirent => dirent.isDirectory())
        .map(dirent => dirent.name);
      
      dirs.forEach(dirName => {
        const infoPath = path.join(clonedDir, dirName, 'info.json');
        if (fs.existsSync(infoPath)) {
          try {
            const info = JSON.parse(fs.readFileSync(infoPath, 'utf8'));
            const relativePath = path.join(dirName);
            info.files = {
              simple: `/cloned/${relativePath}/simple.html`,
              full: `/cloned/${relativePath}/full.html`,
              info: `/cloned/${relativePath}/info.json`
            };
            pages.push(info);
          } catch (error) {
            console.error(`读取${infoPath}失败: ${error.message}`);
          }
        }
      });
    }
  } catch (error) {
    console.error(`获取克隆页面列表失败: ${error.message}`);
  }
  
  res.json(pages);
});

// 启动服务器
app.listen(PORT, '0.0.0.0', () => {
  console.log(`\n🚀 网页克隆工具已启动！`);
  console.log(`🌐 访问地址: http://0.0.0.0:${PORT}`);
  console.log(`📋 本地访问: http://localhost:${PORT}`);
  console.log(`📁 克隆页面保存目录: ${path.join(__dirname, 'cloned_pages')}`);
  console.log(`\n按 Ctrl+C 停止服务器`);
});