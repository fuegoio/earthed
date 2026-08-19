"use client"

import { useState } from "react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import { toast } from "sonner"
import { Check, Loader2, ShieldCheck, X } from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Logo } from "@/components/logo"
import { env } from "@/lib/env"

type State = "idle" | "submitting" | "approved" | "denied" | "error"

export function DeviceConfirm() {
  const searchParams = useSearchParams()
  const userCode = searchParams.get("user_code") ?? ""
  const [state, setState] = useState<State>("idle")
  const [manualCode, setManualCode] = useState(userCode)
  const [showManual, setShowManual] = useState(!userCode)

  const code = (showManual ? manualCode : userCode).trim().toUpperCase()

  async function confirm(deny: boolean) {
    if (!code) {
      toast.error("Please enter your confirmation code")
      return
    }
    setState("submitting")
    try {
      const res = await fetch(
        `${env.NEXT_PUBLIC_EARTHED_API_URL}/api/auth/device/confirm`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ user_code: code, deny }),
          credentials: "include",
        }
      )

      if (!res.ok) {
        const body = await res.json().catch(() => null)
        const detail =
          (body && body.detail) ||
          (body && body.title) ||
          res.statusText ||
          "Request failed"
        if (res.status === 404) {
          setState("error")
          toast.error("Code not found or expired")
        } else if (res.status === 409) {
          setState("error")
          toast.error("This code has already been used")
        } else if (res.status === 401) {
          setState("error")
          toast.error("Please sign in first")
        } else {
          toast.error(detail)
          setState("error")
        }
        return
      }

      setState(deny ? "denied" : "approved")
    } catch {
      setState("error")
      toast.error("Network error")
    }
  }

  if (state === "approved") {
    return (
      <ResultCard
        icon={<Check className="size-8" />}
        title="Approved"
        description="You can close this page. Return to your terminal to finish logging in."
      />
    )
  }

  if (state === "denied") {
    return (
      <ResultCard
        icon={<X className="size-8" />}
        title="Denied"
        description="The login request was denied. You can close this page."
      />
    )
  }

  return (
    <div className="w-full max-w-sm">
      <Link
        href="/"
        className="mb-8 flex items-center justify-center gap-2 font-serif text-2xl font-bold"
      >
        <Logo className="size-8" />
        Earthed
      </Link>
      <Card>
        <CardHeader className="items-center text-center">
          <ShieldCheck className="size-10 text-primary" />
          <CardTitle className="text-xl">Approve CLI login</CardTitle>
          <CardDescription>
            Confirm that this code matches the one shown in your terminal.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {showManual ? (
            <div className="flex flex-col gap-2">
              <input
                className="w-full rounded-md border bg-background px-3 py-3 text-center font-mono text-2xl tracking-widest uppercase"
                placeholder="PLN-XXXX-XXXX"
                value={manualCode}
                onChange={(e) => setManualCode(e.target.value)}
                autoFocus
                maxLength={32}
              />
            </div>
          ) : (
            <div className="rounded-md border bg-muted px-3 py-4 text-center">
              <code className="font-mono text-2xl tracking-widest">
                {userCode}
              </code>
            </div>
          )}

          {!showManual && (
            <button
              onClick={() => setShowManual(true)}
              className="text-center text-sm text-muted-foreground hover:underline"
            >
              Enter a different code
            </button>
          )}

          <div className="flex gap-3">
            <Button
              onClick={() => confirm(false)}
              disabled={state === "submitting" || !code}
              className="flex-1"
            >
              {state === "submitting" && <Loader2 className="size-4 animate-spin" />}
              Approve
            </Button>
            <Button
              variant="outline"
              onClick={() => confirm(true)}
              disabled={state === "submitting" || !code}
              className="flex-1"
            >
              Deny
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function ResultCard({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode
  title: string
  description: string
}) {
  return (
    <div className="w-full max-w-sm">
      <Link
        href="/"
        className="mb-8 flex items-center justify-center gap-2 font-serif text-2xl font-bold"
      >
        <Logo className="size-8" />
        Earthed
      </Link>
      <Card>
        <CardHeader className="items-center text-center">
          <div className="mx-auto mb-2 flex size-16 items-center justify-center rounded-full bg-primary/10 text-primary">
            {icon}
          </div>
          <CardTitle className="text-xl">{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
      </Card>
    </div>
  )
}
