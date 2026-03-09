const puppeteer = require('puppeteer');
const path = require('path');
const fs = require('fs');

const outputDir = path.join(__dirname, '../docs/images/docs');

// 确保输出目录存在
if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir, { recursive: true });
}

const pages = [
  { url: 'https://yi-nology.github.io/git-manage-service/', name: 'docs-home', title: '文档首页' },
  { url: 'https://yi-nology.github.io/git-manage-service/getting-started', name: 'docs-getting-started', title: '快速开始' },
  { url: 'https://yi-nology.github.io/git-manage-service/features/repo', name: 'docs-repo', title: '仓库管理' },
  { url: 'https://yi-nology.github.io/git-manage-service/features/sync', name: 'docs-sync', title: '同步任务' },
  { url: 'https://yi-nology.github.io/git-manage-service/features/spec-editor', name: 'docs-spec-editor', title: 'Spec 编辑器' },
  { url: 'https://yi-nology.github.io/git-manage-service/features/patch-manager', name: 'docs-patch-manager', title: 'Patch 管理' },
  { url: 'https://yi-nology.github.io/git-manage-service/deployment/binary', name: 'docs-deployment', title: '部署方案' },
];

async function screenshot() {
  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  console.log('开始截图...');

  for (const page of pages) {
    console.log(`截图: ${page.title} (${page.url})`);
    
    const tab = await browser.newPage();
    await tab.setViewport({ width: 1280, height: 800 });
    
    try {
      await tab.goto(page.url, { waitUntil: 'networkidle2', timeout: 30000 });
      await new Promise(r => setTimeout(r, 2000)); // 等待页面渲染
      
      const outputPath = path.join(outputDir, `${page.name}.png`);
      await tab.screenshot({ path: outputPath, fullPage: false });
      console.log(`  ✓ 保存到: ${outputPath}`);
    } catch (error) {
      console.error(`  ✗ 失败: ${error.message}`);
    }
    
    await tab.close();
  }

  await browser.close();
  console.log('截图完成！');
}

screenshot().catch(console.error);
