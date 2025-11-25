import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  let base = '/';
  if (mode === 'github') {
    base = '/chainqa-offchain-demo/';
  }

  return {
    plugins: [
      vue(),
      vueDevTools(),
    ],
    server: {
      port: 8099, // 设置开发服务器的端口为 8099
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      },
    },
    base
  };
});
