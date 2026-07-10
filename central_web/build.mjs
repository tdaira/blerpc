// Bundle the browser central to dist/ with esbuild.
//   dist/blerpc-web.js  — IIFE, global `BlerpcWeb` (load with a plain <script>)
//   dist/blerpc-web.mjs — ESM (import { BlerpcClient } from './dist/blerpc-web.mjs')
import { build } from 'esbuild';

const common = {
  entryPoints: ['src/index.ts'],
  bundle: true,
  target: ['es2020'],
  logLevel: 'info',
};

await build({
  ...common,
  format: 'iife',
  globalName: 'BlerpcWeb',
  outfile: 'dist/blerpc-web.js',
});

await build({
  ...common,
  format: 'esm',
  outfile: 'dist/blerpc-web.mjs',
});

console.log('Built dist/blerpc-web.js and dist/blerpc-web.mjs');
