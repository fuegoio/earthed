import { redirect } from "next/navigation"
import { getClient, getMe } from "@/lib/planetary"
import { AppShell } from "@/components/app-shell"

export default async function AppLayout({
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

  return <AppShell userEmail={user.email}>{children}</AppShell>
}
