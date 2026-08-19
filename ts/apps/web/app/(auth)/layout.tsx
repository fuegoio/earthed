import Link from "next/link";
import { Logo } from "@/components/logo";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-8 p-6">
      <Link href="/" className="flex items-center gap-2 font-serif text-2xl font-bold">
        <Logo className="size-8" />
        Earthed
      </Link>
      <div className="w-full max-w-sm">{children}</div>
    </div>
  );
}
