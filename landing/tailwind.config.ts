import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          yellow: '#F4B400',
          'yellow-dark': '#DCA000',
          soft: '#FEF4CD',
          cream: '#FFFCEF',
          white: '#FFFFFF',
          espresso: '#38200F',
          brown: '#806B55',
          border: '#EAE1D1',
        },
        // Semantic status colour only — never used as a brand colour.
        success: {
          DEFAULT: '#2F7D4F',
          soft: '#EAF6EE',
        },
      },
      fontFamily: {
        sans: [
          'Poppins',
          'Noto Sans Kannada',
          'ui-sans-serif',
          'system-ui',
          '-apple-system',
          'Segoe UI',
          'sans-serif',
        ],
        kannada: ['Noto Sans Kannada', 'Poppins', 'sans-serif'],
      },
      boxShadow: {
        card: '0 1px 2px rgba(56, 32, 15, 0.04), 0 18px 44px -28px rgba(56, 32, 15, 0.32)',
        lift: '0 22px 55px -26px rgba(56, 32, 15, 0.4)',
        phone: '0 42px 90px -34px rgba(56, 32, 15, 0.55)',
        float: '0 16px 40px -18px rgba(56, 32, 15, 0.3)',
        nav: '0 1px 0 rgba(56, 32, 15, 0.06), 0 12px 32px -26px rgba(56, 32, 15, 0.3)',
      },
      borderRadius: {
        '4xl': '2rem',
        '5xl': '2.75rem',
      },
      maxWidth: {
        '8xl': '88rem',
      },
      fontSize: {
        display: [
          'clamp(2.65rem, 6.2vw, 4.7rem)',
          { lineHeight: '1.02', letterSpacing: '-0.042em' },
        ],
        'display-sm': [
          'clamp(2.1rem, 4.6vw, 3.4rem)',
          { lineHeight: '1.08', letterSpacing: '-0.035em' },
        ],
      },
      keyframes: {
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-9px)' },
        },
      },
      animation: {
        float: 'float 7s ease-in-out infinite',
        'float-slow': 'float 9s ease-in-out infinite',
      },
      backgroundImage: {
        'dot-grid': 'radial-gradient(circle, rgba(56,32,15,0.16) 1px, transparent 1px)',
        'dot-grid-light': 'radial-gradient(circle, rgba(255,252,239,0.22) 1px, transparent 1px)',
      },
    },
  },
  plugins: [],
} satisfies Config;
