import { redirect } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { getClient, getMe } from "@/lib/planetary";

export default async function SubscribeLayout({ children }: { children: React.ReactNode }) {
  const client = await getClient();

  let user;
  try {
    const result = await getMe({ client });
    user = result.data;
  } catch {
    redirect("/login");
  }

  if (!user) redirect("/login");

  return (
    <div className="flex h-svh flex-col">
      <header className="flex h-14 shrink-0 items-center border-b border-border px-4">
        <Link
          href="/"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Back
        </Link>
      </header>
      <main className="flex min-h-0 flex-1">{children}</main>
    </div>
  );
}
