import { redirect } from "next/navigation";
import { env } from "@/lib/env";
import { safeRedirect } from "@/lib/auth";

export default async function SignupPage({
  searchParams,
}: {
  searchParams: Promise<{ redirect?: string }>;
}) {
  const { redirect: redirectTo } = await searchParams;
  const params = new URLSearchParams();
  if (redirectTo) params.set("redirect", safeRedirect(redirectTo));
  const qs = params.toString();
  redirect(`${env.SUNRED_API_URL}/auth/oauth/signup${qs ? `?${qs}` : ""}`);
}
