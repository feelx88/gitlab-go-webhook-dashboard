import Vue from 'vue'
import VueRouter from 'vue-router'
import Pipelines from '../views/Pipelines'

Vue.use(VueRouter)

const routes = [
  {path: '/:namespace', component: Pipelines}
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes: routes
})

export default router
