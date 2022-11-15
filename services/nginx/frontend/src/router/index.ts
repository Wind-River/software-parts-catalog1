import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import Home from '../views/Home.vue'

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/about',
    name: 'About',
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () => import(/* webpackChunkName: "about" */ '../views/About.vue')
  },
  {
    path: '/file',
    name: 'File',
    component: () => import('../views/File.vue')
  },
  {
    path: '/update',
    name: 'Update',
    component: () => import('../views/Update.vue')
  },
  {
    path: '/container',
    name: 'Package Search',
    component: () => import('../views/PackageSearch.vue')
  },
  {
    path: '/container/:id',
    name: 'Package Detail',
    component: () => import('../views/PackageDetail.vue')
  },
  {
    path: '/group',
    name: 'Group Search',
    component: () => import('../views/GroupSearch.vue')
  },
  {
    path: '/group/:id',
    name: 'Group Detail',
    component: () => import('../views/GroupDetail.vue')
  },
  {
    path: '/missing',
    name: 'Missing',
    component: () => import('../views/Missing.vue')
  },
  {
    path: '/archive',
    name: 'Archive',
    component: () => import('../views/Delete.vue')
  },
  {
    path: '/parts',
    name: 'List Parts',
    component: () => import('../views/Parts.vue')
  },
  {
    path: '/parts/:identifier',
    name: 'View Part',
    props: true,
    component: () => import('../views/Part.vue')
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router
