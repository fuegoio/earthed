import type { Preview } from "@storybook/react";
import { useEffect } from "react";

import "../src/styles/globals.css";

const preview: Preview = {
  parameters: {
    controls: { matchers: { color: /(background|color)$/i, date: /Date$/i } },
    layout: "padded",
    backgrounds: {
      options: {
        Light: { name: "Light", value: "oklch(1 0.003 79)" },
        Dark: { name: "Dark", value: "oklch(0.21 0.006 79)" },
      },
    },
  },
  globalTypes: {
    theme: {
      name: "Theme",
      description: "Light / dark token theme",
      defaultValue: "light",
      toolbar: {
        title: "Theme",
        icon: "circlehollow",
        items: [
          { value: "light", icon: "sun", title: "Light" },
          { value: "dark", icon: "moon", title: "Dark" },
        ],
        dynamicTitle: true,
      },
    },
  },
  decorators: [
    (Story, context) => {
      const theme = context.globals.theme as "light" | "dark";
      useEffect(() => {
        document.documentElement.classList.toggle("dark", theme === "dark");
      }, [theme]);
      return <Story />;
    },
  ],
};

export default preview;
