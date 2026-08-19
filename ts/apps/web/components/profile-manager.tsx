"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Loader2, User } from "lucide-react";
import { Skeleton } from "@workspace/ui/components/skeleton";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import { Label } from "@workspace/ui/components/label";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { getClient, getMe, updateMe, deleteMe, unwrap } from "@/lib/earthed";
import { signout } from "@/lib/auth";
import { getApiErrorMessage } from "@/lib/errors";
import { cn } from "@workspace/ui/lib/utils";
import type { User as UserType } from "@/lib/types";

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

const profileSchema = z.object({
  first_name: z.string().max(255, "Name must be 255 characters or fewer"),
  email: z
    .string()
    .min(1, "Email is required")
    .email("Enter a valid email address")
    .max(255, "Email must be 255 characters or fewer"),
});

type ProfileValues = z.infer<typeof profileSchema>;

// ---------------------------------------------------------------------------
// Avatar
// ---------------------------------------------------------------------------

function AvatarDisplay({ email, firstName }: { email: string; firstName?: string }) {
  const initial = (firstName?.trim() || email).charAt(0).toUpperCase();
  return (
    <div
      className={cn(
        "flex size-16 shrink-0 items-center justify-center rounded-full",
        "bg-primary text-xl font-semibold text-primary-foreground",
        "select-none",
      )}
      aria-hidden="true"
    >
      {initial}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Profile form
// ---------------------------------------------------------------------------

function ProfileForm({ user }: { user: UserType }) {
  const queryClient = useQueryClient();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isDirty, isSubmitting },
  } = useForm<ProfileValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      first_name: user.first_name ?? "",
      email: user.email,
    },
  });

  // Sync form when user data changes (e.g. after a successful save)
  useEffect(() => {
    reset({ first_name: user.first_name ?? "", email: user.email });
  }, [user, reset]);

  async function onSubmit(values: ProfileValues) {
    const { error } = await updateMe({
      client: await getClient(),
      body: { first_name: values.first_name, email: values.email },
    });
    if (error) {
      toast.error(getApiErrorMessage(error, "Could not update profile"));
      return;
    }
    await queryClient.invalidateQueries({ queryKey: ["me"] });
    toast.success("Profile updated");
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-5">
      {/* Avatar + identity */}
      <div className="flex items-center gap-4">
        <AvatarDisplay email={user.email} firstName={user.first_name} />
        <div className="min-w-0">
          <p className="truncate font-medium">
            {user.first_name?.trim() || user.email}
          </p>
          <p className="text-sm text-muted-foreground">{user.email}</p>
        </div>
      </div>

      <div className="h-px bg-border" />

      {/* Fields */}
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="profile-name">Display name</Label>
          <Input
            id="profile-name"
            placeholder="Your name"
            autoComplete="given-name"
            aria-invalid={!!errors.first_name}
            {...register("first_name")}
          />
          {errors.first_name && (
            <p className="text-xs text-destructive">{errors.first_name.message}</p>
          )}
        </div>

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="profile-email">Email address</Label>
          <Input
            id="profile-email"
            type="email"
            placeholder="you@example.com"
            autoComplete="email"
            aria-invalid={!!errors.email}
            {...register("email")}
          />
          {errors.email && (
            <p className="text-xs text-destructive">{errors.email.message}</p>
          )}
        </div>
      </div>

      <div className="flex justify-end">
        <Button type="submit" disabled={isSubmitting || !isDirty}>
          {isSubmitting && <Loader2 className="size-4 animate-spin" />}
          Save changes
        </Button>
      </div>
    </form>
  );
}

// ---------------------------------------------------------------------------
// Danger zone
// ---------------------------------------------------------------------------

function DangerZone({ userEmail }: { userEmail: string }) {
  const router = useRouter();

  async function handleDelete() {
    const { error } = await deleteMe({ client: await getClient() });
    if (error) {
      toast.error(getApiErrorMessage(error, "Could not delete account"));
      throw error;
    }
    // Best-effort sign out before redirecting
    await signout();
    router.push("/login");
    router.refresh();
  }

  return (
    <section
      aria-labelledby="danger-zone-heading"
      className="rounded-lg border border-border p-4"
    >
      <h2
        id="danger-zone-heading"
        className="text-sm font-medium text-muted-foreground"
      >
        Danger zone
      </h2>
      <p className="mt-1 text-sm text-muted-foreground">
        Permanently delete your account and all of your data, including feeds,
        folders, entries, and API tokens. This cannot be undone.
      </p>
      <div className="mt-4">
        <ConfirmDialog
          trigger={
            <Button variant="ghost" size="sm" className="text-muted-foreground">
              Delete account
            </Button>
          }
          title="Delete your account?"
          description={`All data for ${userEmail} will be permanently deleted. This cannot be undone.`}
          confirmLabel="Yes, delete my account"
          onConfirm={handleDelete}
        />
      </div>
    </section>
  );
}

// ---------------------------------------------------------------------------
// Skeleton
// ---------------------------------------------------------------------------

function ProfileSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-4">
        <Skeleton className="size-16 rounded-full shrink-0" />
        <div className="flex flex-col gap-2">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-3 w-48" />
        </div>
      </div>
      <div className="h-px bg-border" />
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-9 w-full" />
        </div>
        <div className="flex flex-col gap-1.5">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-9 w-full" />
        </div>
      </div>
      <div className="flex justify-end">
        <Skeleton className="h-9 w-28 rounded-md" />
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main export
// ---------------------------------------------------------------------------

export function ProfileManager() {
  const { data: user, isLoading } = useQuery<UserType>({
    queryKey: ["me"],
    queryFn: async () => unwrap(getMe({ client: await getClient() })),
  });

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-6 sm:px-6">
      <header>
        <h1 className="flex items-center gap-2 font-serif text-2xl font-bold tracking-normal">
          <User className="size-5" />
          Profile
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Manage your display name, email address, and account.
        </p>
      </header>

      <div className="mt-6 flex flex-col gap-8">
        <section aria-label="Profile information">
          {isLoading || !user ? (
            <ProfileSkeleton />
          ) : (
            <ProfileForm user={user} />
          )}
        </section>

        {user && (
          <>
            <div className="h-px bg-border" />
            <DangerZone userEmail={user.email} />
          </>
        )}
      </div>
    </div>
  );
}
