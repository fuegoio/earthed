import Link from "next/link";
import { buttonVariants } from "@workspace/ui/components/button";
import { cn } from "@workspace/ui/lib/utils";

export default function OfflinePage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-4 text-center">
      <h1 className="text-2xl font-semibold">You&apos;re offline</h1>
      <p className="text-muted-foreground">
        Earthed can&apos;t reach the internet right now. Cached pages will still work — come back
        when you&apos;re reconnected.
      </p>
      <Link href="/" className={cn(buttonVariants())}>
        Try again
      </Link>
    </div>
  );
}
