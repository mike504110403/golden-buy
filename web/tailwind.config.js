/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // 金色主題
        gold: {
          50: '#fdfaf3',
          100: '#faf4e6',
          200: '#f5e8c7',
          300: '#efd89f',
          400: '#e8c46f',
          500: '#d4af37', // 主金色
          600: '#b8952f',
          700: '#967827',
          800: '#735c1f',
          900: '#4d3d15',
        },
        // 價格漲跌色
        up: {
          DEFAULT: '#10b981',
          light: '#34d399',
        },
        down: {
          DEFAULT: '#ef4444',
          light: '#f87171',
        },
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'Menlo', 'Monaco', 'Courier New', 'monospace'],
      },
    },
  },
  plugins: [],
}

