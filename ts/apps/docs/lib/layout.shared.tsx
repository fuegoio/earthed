import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";
import { Logo } from "@/components/logo";

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
