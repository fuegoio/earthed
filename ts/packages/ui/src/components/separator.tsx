import { Separator } from "@base-ui/react/separator";
import { cn } from "@workspace/ui/lib/utils";

function SidebarSeparator({ className, ...props }: React.ComponentProps<typeof Separator>) {
  return (
    <Separator orientation="horizontal" className={cn("bg-sidebar-border", className)} {...props} />
  );
}

export { SidebarSeparator };
