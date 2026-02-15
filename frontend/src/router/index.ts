import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/home/HomePage.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'repos',
        name: 'RepoList',
        component: () => import('@/views/repo/RepoListPage.vue'),
        meta: { title: '仓库管理' },
      },
      {
        path: 'repos/:repoKey',
        name: 'RepoDetail',
        component: () => import('@/views/repo/RepoDetailPage.vue'),
        meta: { title: '仓库详情' },
      },
      {
        path: 'repos/:repoKey/branches',
        name: 'BranchList',
        component: () => import('@/views/branch/BranchListPage.vue'),
        meta: { title: '分支管理' },
      },
      {
        path: 'repos/:repoKey/branches/:branchName',
        name: 'BranchDetail',
        component: () => import('@/views/branch/BranchDetailPage.vue'),
        meta: { title: '分支详情' },
      },
      {
        path: 'repos/:repoKey/compare',
        name: 'BranchCompare',
        component: () => import('@/views/branch/BranchComparePage.vue'),
        meta: { title: '分支对比' },
      },
      {
        path: 'repos/:repoKey/sync',
        name: 'RepoSync',
        component: () => import('@/views/sync/SyncTaskPage.vue'),
        meta: { title: '同步任务' },
      },
      {
        path: 'audit',
        name: 'AuditLog',
        component: () => import('@/views/audit/AuditLogPage.vue'),
        meta: { title: '审计日志' },
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/SettingsPage.vue'),
        meta: { title: '系统设置' },
      },
      {
        path: 'settings/ssh-keys',
        name: 'SSHKeys',
        component: () => import('@/views/settings/SSHKeysPage.vue'),
        meta: { title: 'SSH 密钥管理' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const title = to.meta.title as string
  document.title = title ? `${title} - Git Branch Manager` : 'Git Branch Manager'
  next()
})

export default router
