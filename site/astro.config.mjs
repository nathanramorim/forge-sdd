// @ts-check
import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';

export default defineConfig({
  site: 'https://forge-sdd.org',
  output: 'static',
  compressHTML: true,
  integrations: [sitemap()],
  build: {
    format: 'file'
  }
});
