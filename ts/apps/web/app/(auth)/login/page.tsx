import { Suspense } from "react";
import { AuthForm } from "@/components/auth-form";
import { redirectIfAuthenticated } from "@/lib/auth-guard";
import { getOAuthConfig } from "@/lib/auth";

export const metadata = { title: "Sign in" };

const ERROR_MESSAGES: Record<string, string> = {
  oauth_failed: "Login was cancelled or failed. Please try again.",
  internal: "Something went wrong on our side. Please try again.",
  signup_failed: "Could not start signup. Please try again.",
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

  const errorMessage = error ? ERROR_MESSAGES[error] ?? "Login failed. Please try again." : null;
  const { default_pds } = await getOAuthConfig();

  return (
    <Suspense>
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-1.5 text-center">
          <h1 className="font-serif text-2xl font-bold tracking-tight">
            Welcome to Earthed
          </h1>
          <p className="text-sm text-muted-foreground">
            A calm place to read your feeds.
          </p>
        </div>

        {errorMessage && (
          <div
            role="alert"
            className="rounded-lg border border-destructive/40 bg-destructive/8 px-4 py-3 text-sm text-destructive"
          >
            {errorMessage}
          </div>
        )}

        <AuthForm signupAvailable={!!default_pds} />
      </div>
    </Suspense>
  );
}
