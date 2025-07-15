import { createRouter, createWebHistory } from 'vue-router'
import DatabaseView from '../views/DatabaseView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'database',
      component: DatabaseView,
    },
  ],
})

export default router
