/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./**/*tmpl",
  ],
  plugins: [
      require('@tailwindcss/aspect-ratio'),
      require('@tailwindcss/forms'),
  ],
}
