import { redirect } from "next/navigation";
import { env } from "@/lib/env";
import { safeRedirect } from "@/lib/auth";

export default async function SignupPage({
  searchParams,
}: {
  searchParams: Promise<{ redirect?: string }>;
}) {
  const { redirect: redirectTo } = await searchParams;
  // Start the OAuth flow against the default PDS — same as the "Continue with
  // snrd.social" button on the login page. The PDS decides whether to show
  // login or signup.
  const params = new URLSearchParams({ handle: env.NEXT_PUBLIC_SUNRED_DEFAULT_PDS });
  if (redirectTo) params.set("redirect", safeRedirect(redirectTo));
  redirect(`${env.SUNRED_API_URL}/auth/oauth/login?${params}`);
}
