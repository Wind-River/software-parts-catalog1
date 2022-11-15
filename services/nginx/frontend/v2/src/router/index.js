import Vue from 'vue'
import Router from 'vue-router'
import Home from '@/components/Home.vue'
import File from '@/components/File.vue'
import ContainerDetail from '@/components/ContainerDetail.vue'
import ContainerSearch from '@/components/ContainerSearch.vue'
import GroupDetail from '@/components/GroupDetail.vue'
import GroupSearch from '@/components/GroupSearch.vue'
import Missing from '@/components/Missing.vue'
import Delete from '@/components/Delete.vue'
import Test from '@/components/Test.vue'
import FileView from '@/components/FileView.vue'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home
    },
    {
      path: '/file',
      name: 'File',
      component: File
    },
    {
      path: '/container',
      name: 'Container Search',
      component: ContainerSearch
    },
    {
      path: '/container/:id',
      name: 'Container',
      component: ContainerDetail
    },
    {
      path: '/group',
      name: 'Group Search',
      component: GroupSearch
    },
    {
      path: '/group/:id',
      name: 'Group',
      component: GroupDetail
    },
    {
      path: '/missing',
      name: 'Missing',
      component: Missing
    },
    {
      path: '/archive',
      name: 'Archive',
      component: Delete
    },
    {
      path: '/test',
      name: 'Test',
      component: Test
    },
    {
      path: '/files/:sha256',
      name: 'View File',
      component: FileView
    }
  ]
})
