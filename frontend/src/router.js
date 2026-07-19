import { createRouter, createWebHistory } from 'vue-router'
import HomeView from './pages/HomeView.vue'

const routes = [
  { path: '/', name: 'home', component: HomeView },
  {
    path: '/swiss-cheese',
    name: 'swiss-cheese',
    // Lazy-loaded so the (heavier) SVG model doesn't weigh down the landing page.
    component: () => import('./pages/SwissCheeseView.vue'),
  },
  {
    path: '/about',
    name: 'about',
    component: () => import('./pages/AboutView.vue'),
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})
