"use client";

import { useState } from "react";
import { useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import { Label } from "@workspace/ui/components/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card";
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
    <Card>
      <CardHeader className="items-center text-center">
        <CardTitle className="text-xl">Sign in</CardTitle>
        <CardDescription>
          Use your{" "}
          <a
            href="https://atproto.com"
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium text-primary hover:underline"
          >
            AT Protocol
          </a>{" "}
          handle to log in. This is likely your Bluesky (
          <code className="text-xs">.bsky.social</code>) or Earthed account.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4" noValidate>
          <div className="flex flex-col gap-2">
            <Label htmlFor="handle">Handle</Label>
            <Input
              id="handle"
              type="text"
              placeholder="alice.bsky.social"
              autoComplete="username"
              aria-invalid={!!errors.handle}
              aria-describedby={errors.handle ? "handle-error" : undefined}
              {...register("handle")}
            />
            {errors.handle && (
              <p id="handle-error" className="text-sm text-destructive">
                {errors.handle.message}
              </p>
            )}
          </div>

          <Button type="submit" disabled={isSubmitting} className="w-full">
            {isSubmitting && <Loader2 className="size-4 animate-spin" />}
            Login
          </Button>
        </form>
      </CardContent>
      <CardFooter className="flex-col gap-2 text-center">
        {signupAvailable ? (
          <button
            type="button"
            onClick={() => signupWithDefaultPDS(redirect)}
            className="text-sm text-muted-foreground hover:text-primary hover:underline"
          >
            Don&apos;t have an account? Create one
          </button>
        ) : (
          <p className="text-sm text-muted-foreground">
            New here? Just enter a handle to create an account on your PDS.
          </p>
        )}
      </CardFooter>
    </Card>
  );
}
