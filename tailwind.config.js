module.exports = {
  content: [
    "./web/templates/**/*.html",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        heading: ['Work Sans', 'system-ui', '-apple-system', 'sans-serif'],
      },
    },
    screens: {
      'xs': '306px',
    },
  },
  plugins: [],
}
