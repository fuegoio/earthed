"use client"

import { useEffect } from "react"
import { AlertCircle } from "lucide-react"
import { Button } from "@workspace/ui/components/button"

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <div className="flex flex-col items-center justify-center gap-4 px-6 py-20 text-center">
      <div className="flex size-12 items-center justify-center rounded-xl bg-destructive/10">
        <AlertCircle className="size-6 text-destructive" />
      </div>
      <div>
        <h1 className="font-serif text-lg font-bold tracking-tight">
          Something went wrong
        </h1>
        <p className="mt-1 max-w-sm text-sm text-muted-foreground">
          An unexpected error occurred while loading this page.
        </p>
      </div>
      <Button onClick={reset} size="sm">
        Try again
      </Button>
    </div>
  )
}
