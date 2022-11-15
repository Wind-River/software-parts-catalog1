// Styles
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

// Vuetify
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

export default createVuetify({
  theme: {
    themes: {
      light: {
        colors: {
          primary: "#00ADA4",
          secondary: "#AD3C00",
          accent: "#ff5722",
          error: "#f44336",
          warning: "#ffc107",
          info: "#ffeb3b",
          success: "#4caf50",
        },
      },
    },
  },
}
  // https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
)
