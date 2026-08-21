import { Suspense } from "react";
import { AuthForm } from "@/components/auth-form";
import { redirectIfAuthenticated } from "@/lib/auth-guard";

export const metadata = { title: "Sign in" };

const ERROR_MESSAGES: Record<string, string> = {
  oauth_failed: "Login was cancelled or failed. Please try again.",
  internal: "Something went wrong on our side. Please try again.",
};

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ redirect?: string; error?: string }>;
}) {
  const { redirect: redirectTo, error } = await searchParams;
  // If the visitor already has a valid session, skip the form and go to the
  // redirect target (or home). 4xx/5xx from the API → render the form.
  await redirectIfAuthenticated(redirectTo);

  const errorMessage = error ? ERROR_MESSAGES[error] ?? "Login failed." : null;

  return (
    <Suspense>
      <div className="flex flex-col gap-4">
        {errorMessage && (
          <div className="rounded-md border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {errorMessage}
          </div>
        )}
        <AuthForm />
      </div>
    </Suspense>
  );
}
