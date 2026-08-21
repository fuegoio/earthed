"use client";

import { useState } from "react";
import { useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@workspace/ui/components/field";
import { loginWithHandle, safeRedirect, signupWithDefaultPDS } from "@/lib/auth";
import { handleSchema, type HandleValues } from "@/lib/schemas";

export function AuthForm({ signupAvailable = false }: { signupAvailable?: boolean }) {
  const searchParams = useSearchParams();
  const redirect = safeRedirect(searchParams.get("redirect"));
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<HandleValues>({
    resolver: zodResolver(handleSchema),
  });

  function onSubmit(values: HandleValues) {
    setIsSubmitting(true);
    // Full-page navigation to the API's OAuth login endpoint. The API
    // redirects to the user's PDS to approve, then back here with a session.
    loginWithHandle(values.handle.trim(), redirect);
  }

  return (
    <div className="flex flex-col gap-6">
      <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4" noValidate>
        <Field data-invalid={!!errors.handle}>
          <FieldLabel htmlFor="handle">Handle</FieldLabel>
          <Input
            id="handle"
            type="text"
            placeholder="alice.bsky.social"
            autoComplete="username"
            autoFocus
            className="h-11 text-base"
            aria-invalid={!!errors.handle}
            {...register("handle")}
          />
          {errors.handle ? (
            <FieldError>{errors.handle.message}</FieldError>
          ) : (
            <FieldDescription>
              Use your{" "}
              <a
                href="https://atproto.com"
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-foreground hover:underline"
              >
                AT Protocol
              </a>{" "}
              handle to log in. If you&apos;re unsure, this is likely your Bluesky (
              <code className="text-xs">.bsky.social</code>) account.
            </FieldDescription>
          )}
        </Field>

        <Button type="submit" size="lg" disabled={isSubmitting} className="w-full">
          {isSubmitting && <Loader2 className="size-4 animate-spin" />}
          Login
        </Button>
      </form>

      {signupAvailable && (
        <p className="text-center text-sm text-muted-foreground">
          Don&apos;t have an account?{" "}
          <button
            type="button"
            onClick={() => signupWithDefaultPDS(redirect)}
            className="font-medium text-foreground hover:underline"
          >
            Create an account
          </button>{" "}
          now.
        </p>
      )}
    </div>
  );
}
