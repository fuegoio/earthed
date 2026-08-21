import Link from "next/link";
import { Logo } from "@/components/logo";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-8 p-6">
      <Link href="/" className="flex flex-col items-center gap-3">
        <Logo className="size-14" />
        <span className="font-serif text-2xl font-bold">Sunred</span>
      </Link>
      <div className="w-full max-w-sm">{children}</div>
    </div>
  );
}
