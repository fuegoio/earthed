import { redirect } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { getClient, getMe } from "@/lib/sunred";
import { apiErrorStatus, getApiErrorMessage, isClientError } from "@/lib/errors";
import { ApiError } from "@/components/api-error";

export default async function SubscribeLayout({ children }: { children: React.ReactNode }) {
  const client = await getClient();

  let user: Awaited<ReturnType<typeof getMe>>["data"];
  let err: unknown;
  try {
    const result = await getMe({ client });
    if (result.error) {
      err = result.error;
    } else {
      user = result.data;
    }
  } catch (e) {
    err = e;
  }

  if (err !== undefined && isClientError(err)) {
    redirect("/login");
  }
  if (user === undefined && err === undefined) {
    redirect("/login");
  }

  if (user === undefined) {
    const status = apiErrorStatus(err) ?? 503;
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
        <main className="flex min-h-0 flex-1 items-center justify-center p-8">
          <ApiError
            status={status}
            message={getApiErrorMessage(err, "The API is unreachable. Please try again later.")}
            className="max-w-md"
          />
        </main>
      </div>
    );
  }

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
