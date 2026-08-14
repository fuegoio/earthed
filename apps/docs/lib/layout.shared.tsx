import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";

export function baseOptions(props: BaseLayoutProps = {}): BaseLayoutProps {
  return {
    nav: {
      title: "Planetary",
    },
    ...props,
  };
}
