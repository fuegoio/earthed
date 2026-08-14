"use client"

import { useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { ArrowLeft, Loader2, ListChecks } from "lucide-react"
import { Button, buttonVariants } from "@workspace/ui/components/button"
import { Input } from "@workspace/ui/components/input"
import { Label } from "@workspace/ui/components/label"
import { Textarea } from "@workspace/ui/components/textarea"
import { getClient, createFeedList } from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"

export default function NewFeedListPage() {
  const router = useRouter()
  const [title, setTitle] = useState("")
  const [description, setDescription] = useState("")
  const [isPublic, setIsPublic] = useState(false)
  const [pending, setPending] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!title.trim()) return
    setPending(true)
    try {
      const { data, error } = await createFeedList({
        client: await getClient(),
        body: {
          title: title.trim(),
          description: description.trim(),
          is_public: isPublic,
        },
      })
      if (error) throw error
      toast.success("Feed list created")
      router.push(`/lists/${data!.id}`)
      router.refresh()
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not create feed list"))
      setPending(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-8 sm:px-6">
      <Link
        href="/lists"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="size-3.5" />
        Back to lists
      </Link>

      <div className="flex flex-col gap-3">
        <div className="flex size-12 items-center justify-center rounded-xl bg-primary/10">
          <ListChecks className="size-6 text-primary" />
        </div>
        <h1 className="font-serif text-2xl font-bold tracking-tight">
          New feed list
        </h1>
        <p className="max-w-md text-sm text-muted-foreground">
          A feed list is a curated collection of RSS feeds. Add feeds, then
          share it publicly so others can follow and import them in one click.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="mt-8 flex flex-col gap-5">
        <div className="flex flex-col gap-2">
          <Label htmlFor="title">Title</Label>
          <Input
            id="title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Best engineering blogs"
            maxLength={255}
            autoFocus
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="description">Description (optional)</Label>
          <Textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="What is this list about?"
            maxLength={2000}
            rows={3}
          />
        </div>

        <label className="flex cursor-pointer items-start gap-3 rounded-md border border-border p-3">
          <input
            type="checkbox"
            checked={isPublic}
            onChange={(e) => setIsPublic(e.target.checked)}
            className="mt-0.5 size-4 accent-primary"
          />
          <span className="flex flex-col">
            <span className="text-sm font-medium">Make this list public</span>
            <span className="text-sm text-muted-foreground">
              Public lists can be discovered and followed by other users. You
              can change this later.
            </span>
          </span>
        </label>

        <div className="flex gap-2">
          <Button type="submit" disabled={pending || !title.trim()}>
            {pending && <Loader2 className="size-4 animate-spin" />}
            Create list
          </Button>
          <Link
            href="/lists"
            className={buttonVariants({ variant: "ghost" })}
          >
            Cancel
          </Link>
        </div>
      </form>
    </div>
  )
}
