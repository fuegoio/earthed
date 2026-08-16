import { User } from "lucide-react";
import { getClient, getMe } from "@/lib/planetary";

export const metadata = { title: "Profile" };

export default async function ProfilePage() {
  const client = await getClient();
  const { data: user } = await getMe({ client });

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-6 sm:px-6">
      <header className="flex items-center gap-2">
        <User className="size-5 text-muted-foreground" />
        <h1 className="font-serif text-2xl font-bold tracking-tight">Profile</h1>
      </header>
      <p className="mt-2 text-sm text-muted-foreground">
        Your account details and session information.
      </p>
      <div className="mt-6 rounded-lg border border-border p-4">
        <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Email</p>
        <p className="mt-1 text-sm">{user?.email}</p>
      </div>
    </div>
  );
}
