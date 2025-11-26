import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

// Provide a minimal localStorage polyfill so plugins that expect browser APIs
// (e.g. vue devtools) can run while Vite loads the config inside Node.
const ensureLocalStorage = () => {
  const storage = new Map();
  return {
    getItem(key) {
      return storage.has(key) ? storage.get(key) : null;
    },
    setItem(key, value) {
      storage.set(key, String(value));
    },
    removeItem(key) {
      storage.delete(key);
    },
    clear() {
      storage.clear();
    },
    key(index) {
      return Array.from(storage.keys())[index] ?? null;
    },
    get length() {
      return storage.size;
    },
  };
};

if (
  typeof globalThis.localStorage === 'undefined' ||
  typeof globalThis.localStorage.getItem !== 'function'
) {
  globalThis.localStorage = ensureLocalStorage();
}

// https://vite.dev/config/
export default defineConfig(async ({ mode }) => {
  const isDev = mode === 'development';
  const base = mode === 'github' ? '/chainqa-offchain-demo/' : '/';
  const devPlugins = [];

  if (isDev) {
    const { default: vueDevTools } = await import('vite-plugin-vue-devtools');
    devPlugins.push(vueDevTools());
  }

  return {
    plugins: [
      vue(),
      ...devPlugins,
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
