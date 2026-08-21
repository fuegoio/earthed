"use client";

import { useState } from "react";
import { useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import { Label } from "@workspace/ui/components/label";
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
        <div className="flex flex-col gap-2">
          <Label htmlFor="handle" className="text-foreground">
            Bluesky handle
          </Label>
          <Input
            id="handle"
            type="text"
            placeholder="alice.bsky.social"
            autoComplete="username"
            autoFocus
            className="h-11 text-base"
            aria-invalid={!!errors.handle}
            aria-describedby={errors.handle ? "handle-error" : "handle-hint"}
            {...register("handle")}
          />
          {errors.handle ? (
            <p id="handle-error" className="text-sm text-destructive">
              {errors.handle.message}
            </p>
          ) : (
            <p id="handle-hint" className="text-sm text-muted-foreground">
              The same account you use in the Bluesky app.
            </p>
          )}
        </div>

        <Button type="submit" size="lg" disabled={isSubmitting} className="w-full">
          {isSubmitting && <Loader2 className="size-4 animate-spin" />}
          Continue
        </Button>
      </form>

      {signupAvailable && (
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-3" aria-hidden="true">
            <span className="h-px flex-1 bg-border" />
            <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              or
            </span>
            <span className="h-px flex-1 bg-border" />
          </div>
          <Button
            type="button"
            variant="outline"
            size="lg"
            className="w-full"
            onClick={() => signupWithDefaultPDS(redirect)}
          >
            Create a new account
          </Button>
        </div>
      )}
    </div>
  );
}
