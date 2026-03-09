module.exports = {
  title: 'Git Manage Service',
  description: 'Git 管理服务文档',
  base: '/git-manage-service/',
  themeConfig: {
    nav: [
      {
        text: '首页',
        link: '/'
      },
      {
        text: '使用指南',
        link: '/usage.md'
      },
      {
        text: 'Spec 编辑器',
        link: '/spec-editor.md'
      },
      {
        text: 'Patch 管理',
        link: '/patch-manager.md'
      },
      {
        text: '部署指南',
        link: '/deployment.md'
      },
      {
        text: 'Webhook',
        link: '/webhook.md'
      },
      {
        text: 'GitHub',
        link: 'https://github.com/yi-nology/git-manage-service'
      }
    ],
    sidebar: {
      '/': [
        '',
        'usage',
        'spec-editor',
        'patch-manager',
        'deployment',
        'product_manual',
        'webhook'
      ],
      '/dev-notes/': [
        '',
        'BROWSER_TEST_REPORT',
        'BUGFIX_ISDIR_FIELD',
        'BUGFIX_TREE_STRUCTURE',
        'DEV_SERVER_STATUS',
        'E2E_TEST_REPORT',
        'OPTIMIZATION_PLAN',
        'OPTIMIZATION_REPORT',
        'SPEC_EDITOR_AUTO_GUIDE',
        'SPEC_EDITOR_INIT_FEATURE',
        'SPEC_EDITOR_PLAN',
        'SPEC_EDITOR_TEST_REPORT',
        'SPEC_EDITOR_TEST_REPORT_FINAL',
        'SSH_TROUBLESHOOTING',
        'SYNC_REDESIGN_PLAN',
        'TEST_REPORT',
        'TODO-Git管理服务整体优化-20260308'
      ]
    },
    lastUpdated: '最后更新',
    search: true,
    searchMaxSuggestions: 10
  },
  markdown: {
    lineNumbers: true
  }
}