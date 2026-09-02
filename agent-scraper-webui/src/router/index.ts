import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/lib/api'
import ExpensesView from '@/views/ExpensesView.vue'
import ForwardingView from '@/views/ForwardingView.vue'
import JobsView from '@/views/JobsView.vue'
import LoginView from '@/views/LoginView.vue'
import SourcesView from '@/views/SourcesView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    { path: '/', redirect: '/jobs' },
    { path: '/jobs', name: 'jobs', component: JobsView, meta: { requiresAuth: true } },
    { path: '/expenses', name: 'expenses', component: ExpensesView, meta: { requiresAuth: true } },
    { path: '/sources', name: 'sources', component: SourcesView, meta: { requiresAuth: true } },
    { path: '/forwarding', name: 'forwarding', component: ForwardingView, meta: { requiresAuth: true } },
  ],
})

router.beforeEach((to) => {
  const token = getToken()
  if (to.meta.public) {
    if (to.path === '/login' && token) return '/jobs'
    return true
  }
  if (to.meta.requiresAuth && !token) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
