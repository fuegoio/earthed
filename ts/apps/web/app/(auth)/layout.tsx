import Link from "next/link"
import { Rss } from "lucide-react"

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-8 p-6">
      <Link
        href="/"
        className="flex items-center gap-2 font-serif text-2xl font-bold"
      >
        <Rss className="size-7 text-primary" />
        Planetary
      </Link>
      <div className="w-full max-w-sm">{children}</div>
    </div>
  )
}
