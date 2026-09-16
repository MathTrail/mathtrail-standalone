import { defineConfig } from "vite";
import { viteSingleFile } from "vite-plugin-singlefile";

// Spike widget build: one self-contained HTML file, embedded into the Go
// binary via go:embed. No external domains (CSP has none declared).
export default defineConfig({
  plugins: [viteSingleFile()],
  build: {
    outDir: "dist",
    assetsInlineLimit: 100_000_000,
    cssCodeSplit: false,
    target: "es2022",
  },
});
