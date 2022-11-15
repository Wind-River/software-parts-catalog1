<template>
  <app-page>
    <spa-header slot="header" title="TKDB"/>
    <header-bar slot="nav" :routes="routes"/>
    <left-side-bar slot="lsidebar"/>
    <right-side-bar slot="rsidebar"/>
    <content>
      <router-view/>
    </content>
  </app-page>
</template>

<script>
import AppPage from '@/frame/SPA.vue'
import SpaHeader from '@/frame/Header.vue'
import HeaderBar from '@/frame/HeaderBar.vue'
import LeftSideBar from '@/frame/LeftSideBar.vue'
import RightSideBar from '@/frame/RightSideBar.vue'
import Content from '@/frame/Content.vue'
import Footer from '@/frame/Footer.vue'
import router from '@/router'

export default {
  name: 'App',
  data: () => {
    return {
      title: 'TKDB',
      router: router,
      routes: [
        {
          path: '/file',
          title: 'File'
        },
        {
          path: '/container',
          title: 'Packages'
        },
        {
          path: '/group',
          title: 'Groups'
        },
        {
          path: '/missing',
          title: 'Missing'
        },
        {
          path: '/archive',
          title: 'Archive'
        }
        // {
        //   path: 'test',
        //   title: 'Test'
        // }
      ]
    }
  },
  components: {
    AppPage,
    SpaHeader,
    HeaderBar,
    LeftSideBar,
    RightSideBar,
    Content,
    Footer
  }
}
</script>

<style>
  body {
    margin: 0;
  }
</style>
