const axios = require('axios');
const cheerio = require('cheerio');
const fs = require('fs');
const path = require('path');

async function cloneWebPage(url) {
  try {
    console.log(`正在访问 URL: ${url}`);
    const response = await axios.get(url, {
      headers: {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
      }
    });

    const html = response.data;
    const $ = cheerio.load(html);

    // 提取关键信息
    const title = $('title').text().trim() || '无标题';
    const header = $('header').html() || '';
    const body = $('body').html() || '';
    const head = $('head').html() || '';

    console.log(`成功获取网页: ${title}`);

    // 创建保存目录
    const saveDir = path.join(__dirname, 'cloned_pages', encodeURIComponent(url.replace(/[^a-zA-Z0-9]/g, '_')));
    if (!fs.existsSync(saveDir)) {
      fs.mkdirSync(saveDir, { recursive: true });
    }

    // 保存完整HTML
    const fullHtmlPath = path.join(saveDir, 'full.html');
    fs.writeFileSync(fullHtmlPath, html);
    console.log(`完整HTML已保存到: ${fullHtmlPath}`);

    // 保存提取的信息
    const extractedInfo = {
      url,
      title,
      timestamp: new Date().toISOString(),
      head,
      header,
      body
    };

    const infoPath = path.join(saveDir, 'info.json');
    fs.writeFileSync(infoPath, JSON.stringify(extractedInfo, null, 2));
    console.log(`提取的信息已保存到: ${infoPath}`);

    // 保存简化版HTML
    const simpleHtml = `<!DOCTYPE html>
<html>
<head>
  <title>${title}</title>
  ${head}
</head>
<body>
  ${header}
  ${body}
</body>
</html>`;

    const simplePath = path.join(saveDir, 'simple.html');
    fs.writeFileSync(simplePath, simpleHtml);
    console.log(`简化HTML已保存到: ${simplePath}`);

    return {
      success: true,
      title,
      saveDir,
      files: {
        full: fullHtmlPath,
        simple: simplePath,
        info: infoPath
      }
    };
  } catch (error) {
    console.error(`克隆失败: ${error.message}`);
    return {
      success: false,
      error: error.message
    };
  }
}

// 命令行执行
if (require.main === module) {
  const url = process.argv[2];
  if (!url) {
    console.error('请提供URL参数，例如: node clone.js https://example.com');
    process.exit(1);
  }
  cloneWebPage(url);
}

module.exports = cloneWebPage;