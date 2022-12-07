// Styles
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

// Vuetify
import { createVuetify } from 'vuetify'

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
