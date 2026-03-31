import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './app/**/*.{ts,tsx}',
    './components/**/*.{ts,tsx}',
    './lib/**/*.{ts,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        border: '#dae3ea',
        background: '#f6f0e8',
        foreground: '#173042',
        card: '#ffffff',
        muted: '#eef2f4',
        accent: '#0f7d98',
        warning: '#ca9546',
        success: '#1f9d72',
        danger: '#d95c5c',
      },
      boxShadow: {
        panel: '0 16px 40px rgba(23, 48, 66, 0.08)',
      },
      borderRadius: {
        xl: '1rem',
        '2xl': '1.5rem',
      },
      fontFamily: {
        sans: ['"IBM Plex Sans"', '"Noto Sans SC"', 'sans-serif'],
      },
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        'fade-out': {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        'zoom-in-95': {
          '0%': { transform: 'translate(-50%, -50%) scale(0.95)' },
          '100%': { transform: 'translate(-50%, -50%) scale(1)' },
        },
        'zoom-out-95': {
          '0%': { transform: 'translate(-50%, -50%) scale(1)' },
          '100%': { transform: 'translate(-50%, -50%) scale(0.95)' },
        },
        'slide-in-from-left-1/2': {
          '0%': { transform: 'translateX(-50%)' },
          '100%': { transform: 'translateX(0)' },
        },
        'slide-out-to-left-1/2': {
          '0%': { transform: 'translateX(0)' },
          '100%': { transform: 'translateX(-50%)' },
        },
        'slide-in-from-top-[48%]': {
          '0%': { transform: 'translateY(-48%)' },
          '100%': { transform: 'translateY(0)' },
        },
        'slide-out-to-top-[48%]': {
          '0%': { transform: 'translateY(0)' },
          '100%': { transform: 'translateY(-48%)' },
        },
      },
      animation: {
        'fade-in': 'fade-in 0.2s ease-out',
        'fade-out': 'fade-out 0.2s ease-in',
        'zoom-in-95': 'zoom-in-95 0.2s ease-out',
        'zoom-out-95': 'zoom-out-95 0.2s ease-in',
        'slide-in-from-left-1/2': 'slide-in-from-left-1/2 0.2s ease-out',
        'slide-out-to-left-1/2': 'slide-out-to-left-1/2 0.2s ease-in',
        'slide-in-from-top-[48%]': 'slide-in-from-top-[48%] 0.2s ease-out',
        'slide-out-to-top-[48%]': 'slide-out-to-top-[48%] 0.2s ease-in',
      },
    },
  },
  plugins: [],
}

export default config
