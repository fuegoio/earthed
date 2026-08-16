import type { BaseLayoutProps, LayoutTab } from "fumadocs-ui/layouts/shared";
import { Logo } from "@/components/logo";

const tabs: LayoutTab[] = [
  { title: "Docs", url: "/docs" },
  { title: "Self-Hosting", url: "/self-hosting" },
  { title: "API Reference", url: "/api-reference" },
];

export function baseOptions(props: BaseLayoutProps = {}): BaseLayoutProps {
  return {
    nav: {
      title: (
        <span className="flex items-center gap-2 font-semibold">
          <Logo className="size-5" />
          Planetary
        </span>
      ),
    },
    ...props,
  };
}

export { tabs };
