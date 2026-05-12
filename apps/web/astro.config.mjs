// astro.config.mjs — Configuração do Astro 5.
//
// Tailwind 4 usa o plugin Vite nativo (@tailwindcss/vite) em vez do
// integration do Astro — é mais rápido e não precisa de tailwind.config.js.
// A configuração do tema fica toda em CSS com @theme.
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  integrations: [
    react(), // islands React hidratam no browser com client:load / client:visible
  ],
  vite: {
    plugins: [
      tailwindcss(), // Tailwind 4 como plugin Vite — sem tailwind.config.js
    ],
    server: {
      proxy: {
        // Em dev, /api/* é proxied para o Go backend
        // Evita CORS e permite usar a mesma origem
        '/api': {
          target: process.env.VITE_API_URL ?? 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  },
});
