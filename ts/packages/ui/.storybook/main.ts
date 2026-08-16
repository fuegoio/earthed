import type { StorybookConfig } from "@storybook/react-vite";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const dir = dirname(fileURLToPath(import.meta.url));

const config: StorybookConfig = {
  stories: ["../src/**/*.stories.@(ts|tsx)"],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
  viteFinal: async (cfg) => {
    cfg.resolve ??= {};
    cfg.resolve.alias = {
      ...cfg.resolve.alias,
      // Resolve the workspace alias to the package source so stories can
      // import @workspace/ui/* (components/button, lib/utils, ...) without
      // relying on Node's package "exports" map, which Vite aliases skip.
      "@workspace/ui": join(dir, "..", "src"),
    };
    return cfg;
  },
  docs: {
    autodocs: "tag",
  },
};

export default config;
