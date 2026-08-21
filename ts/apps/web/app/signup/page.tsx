import { redirect } from "next/navigation";

export default async function SignupPage({
  searchParams,
}: {
  searchParams: Promise<{ redirect?: string }>;
}) {
  const { redirect: redirectTo } = await searchParams;
  redirect(redirectTo ? `/login?redirect=${encodeURIComponent(redirectTo)}` : "/login");
}
