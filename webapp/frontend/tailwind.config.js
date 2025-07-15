/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "**/*.{go,js,templ,html}",
    "!**/node_modules"
  ],
  theme: {
    extend: {
      colors: {
        primary: "#bb86fc",
        variant: "#3700b3",
        secondary: "#03dac6",
        background: "#121212",
        surface: "#121212",
        warning: "#fffc99",
        error: "#cf6679",
        on: {
          primary: "#000000",
          secondary: "#000000",
          background: "#ffffff",
          surface: "#ffffff",
          waning: "#000000",
          error: "#000000",
        }
      }
    },
  },
  plugins: [],
}

