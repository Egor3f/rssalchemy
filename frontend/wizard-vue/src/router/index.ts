import { createRouter, createWebHistory } from 'vue-router'
import WizardPage from "@/pages/WizardPage.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'wizard',
      component: WizardPage,
    },
  ],
})

export default router
