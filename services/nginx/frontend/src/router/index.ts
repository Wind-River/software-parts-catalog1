import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router"
import Home from "../views/Home.vue"

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: "Home",
    component: Home,
  },
  {
    path: "/about",
    name: "About",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/About.vue"),
  },
  {
    path: "/file",
    name: "File",
    component: () => import("../views/File.vue"),
  },
  {
    path: "/file/:id",
    name: "PartListAdd",
    component: () => import("../views/File.vue")
  },
  {
    path: "/update",
    name: "Update",
    component: () => import("../views/Update.vue"),
  },
  {
    path: "/container/:id",
    name: "Package Detail",
    component: () => import("../views/PackageDetail.vue"),
  },
  {
    path: "/partlists",
    name: "PartLists",
    component: () => import("../views/PartListBrowser.vue"),
  },
  {
    path: "/partlist/:id",
    name: "Part List Detail",
    component: () => import("../views/PartListDetail.vue"),
  },
  {
    path: "/profile/:id/:key",
    name: "Profile Detail",
    component: () => import("../views/ProfileDetail.vue")
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
