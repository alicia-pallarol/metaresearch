import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import HomeView from '@/pages/HomeView.vue'

// The map is the landing page and is loaded eagerly; everything else is split
// out so the first paint carries only what it needs.
const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/agenda/:id', name: 'agenda', component: () => import('@/pages/AgendaView.vue') },
  { path: '/about', name: 'about', component: () => import('@/pages/AboutView.vue') },
  { path: '/methodology', name: 'methodology', component: () => import('@/pages/MethodologyView.vue') },
  { path: '/privacy', name: 'privacy', component: () => import('@/pages/PrivacyView.vue') },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
})
