import { createApp } from "vue"
import { createPinia } from "pinia"
import App from "./App.vue"
import router from "./router"
import vuetify from "./plugins/vuetify"
import urql from "@urql/vue"
import { loadFonts } from "./plugins/webfontloader"
import { multipartFetchExchange } from "@urql/exchange-multipart-fetch"

loadFonts()
const pinia = createPinia()

createApp(App)
  .use(router)
  .use(vuetify)
  .use(pinia)
  .use(urql, {
    url: "/api/graphql",
    exchanges: [multipartFetchExchange],
  })
  .mount("#app")
