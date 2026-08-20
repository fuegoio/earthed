import { Suspense } from "react";
import { AuthForm } from "@/components/auth-form";
import { redirectIfAuthenticated } from "@/lib/auth-guard";

export const metadata = { title: "Sign in" };

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ redirect?: string }>;
}) {
  const { redirect: redirectTo } = await searchParams;
  // If the visitor already has a valid session, skip the form and go to the
  // redirect target (or home). 4xx/5xx from the API → render the form.
  await redirectIfAuthenticated(redirectTo);

  return (
    <Suspense>
      <AuthForm mode="signin" />
    </Suspense>
  );
}
