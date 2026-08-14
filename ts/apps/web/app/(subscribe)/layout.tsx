import { redirect } from "next/navigation"
import Link from "next/link"
import { getClient, getMe } from "@/lib/planetary"
import { Logo } from "@/components/logo"

export default async function SubscribeLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const client = await getClient()

  let user
  try {
    const result = await getMe({ client })
    user = result.data
  } catch {
    redirect("/login")
  }

  if (!user) redirect("/login")

  return (
    <div className="flex min-h-svh flex-col">
      <header className="flex h-14 shrink-0 items-center border-b border-border px-4">
        <Link
          href="/"
          className="flex items-center gap-2 font-serif text-lg font-bold"
        >
          <Logo className="size-5" />
          Planetary
        </Link>
      </header>
      <main className="flex-1">{children}</main>
    </div>
  )
}
