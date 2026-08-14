"use client"

import { useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import {
  ArrowLeft,
  Loader2,
  Rss,
  ExternalLink,
  Plus,
} from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { Input } from "@workspace/ui/components/input"
import { Label } from "@workspace/ui/components/label"
import {
  getClient,
  listCategories,
  previewFeed,
  createFeed,
  unwrap,
} from "@/lib/planetary"
import { getApiErrorMessage } from "@/lib/errors"
import {
  subscribeFeedSchema,
  type SubscribeFeedValues,
} from "@/lib/schemas"
import type { PreviewFeedBody, Category } from "@/lib/types"
import { cn } from "@workspace/ui/lib/utils"

type Phase = "input" | "preview" | "subscribing"

export default function SubscribeFeedPage() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const [phase, setPhase] = useState<Phase>("input")
  const [preview, setPreview] = useState<PreviewFeedBody | null>(null)
  const [selectedCategoryId, setSelectedCategoryId] = useState<number | "">("")

  const { data: categories } = useQuery<Category[]>({
    queryKey: ["categories"],
    queryFn: async () =>
      unwrap(listCategories({ client: await getClient() })),
  })

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<SubscribeFeedValues>({
    resolver: zodResolver(subscribeFeedSchema),
  })

  async function onSubmit(values: SubscribeFeedValues) {
    setPhase("preview")
    setPreview(null)
    try {
      const result = await previewFeed({
        client: await getClient(),
        body: { feed_url: values.feed_url },
      })
      if (result.error) {
        throw result.error
      }
      setPreview(result.data)
    } catch (err) {
      setError("feed_url", {
        message: getApiErrorMessage(err, "Could not fetch that feed"),
      })
      setPhase("input")
    }
  }

  async function handleSubscribe() {
    if (!preview) return
    setPhase("subscribing")
    try {
      const categoryId =
        selectedCategoryId === "" ? undefined : Number(selectedCategoryId)
      const { error } = await createFeed({
        client: await getClient(),
        body: {
          feed_url: preview.feed_url,
          ...(categoryId ? { category_id: categoryId } : {}),
        },
      })
      if (error) throw error
      await queryClient.invalidateQueries({ queryKey: ["feeds"] })
      toast.success(`Subscribed to "${preview.title}"`)
      router.push("/")
      router.refresh()
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not subscribe to feed"))
      setPhase("preview")
    }
  }

  function handleBackToInput() {
    setPhase("input")
    setPreview(null)
    setSelectedCategoryId("")
  }

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-8 sm:px-6 sm:py-12">
      <Link
        href="/"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="size-3.5" />
        Back
      </Link>

      {phase === "input" && (
        <InputPhase
          register={register}
          handleSubmit={handleSubmit}
          onSubmit={onSubmit}
          errors={errors}
        />
      )}

      {phase === "preview" && !preview && <PreviewLoading />}

      {phase === "preview" && preview && (
        <PreviewPhase
          preview={preview}
          categories={categories ?? []}
          selectedCategoryId={selectedCategoryId}
          onSelectCategory={setSelectedCategoryId}
          onSubscribe={handleSubscribe}
          onBack={handleBackToInput}
        />
      )}

      {phase === "subscribing" && preview && (
        <SubscribingPhase preview={preview} />
      )}
    </div>
  )
}

// --- Input phase ---

function InputPhase({
  register,
  handleSubmit,
  onSubmit,
  errors,
}: {
  register: ReturnType<typeof useForm<SubscribeFeedValues>>["register"]
  handleSubmit: ReturnType<typeof useForm<SubscribeFeedValues>>["handleSubmit"]
  onSubmit: (values: SubscribeFeedValues) => void
  errors: ReturnType<typeof useForm<SubscribeFeedValues>>["formState"]["errors"]
}) {
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleFormSubmit(values: SubscribeFeedValues) {
    setIsSubmitting(true)
    try {
      await onSubmit(values)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-3">
        <div className="flex size-12 items-center justify-center rounded-xl bg-primary/10">
          <Rss className="size-6 text-primary" />
        </div>
        <h1 className="font-serif text-2xl font-bold tracking-tight">
          Subscribe to a feed
        </h1>
        <p className="max-w-md text-sm text-muted-foreground">
          Paste the URL of any RSS, Atom, or JSON feed. We&apos;ll fetch it and
          show you what you&apos;ll get before you subscribe.
        </p>
      </div>

      <form
        onSubmit={handleSubmit(handleFormSubmit)}
        className="flex flex-col gap-4"
        noValidate
      >
        <div className="flex flex-col gap-2">
          <Label htmlFor="feed_url">Feed URL</Label>
          <Input
            id="feed_url"
            type="url"
            placeholder="https://example.com/feed.xml"
            autoComplete="url"
            spellCheck={false}
            aria-invalid={!!errors.feed_url}
            aria-describedby={
              errors.feed_url ? "feed_url-error" : undefined
            }
            {...register("feed_url")}
          />
          {errors.feed_url && (
            <p id="feed_url-error" className="text-sm text-destructive">
              {errors.feed_url.message}
            </p>
          )}
        </div>

        <Button type="submit" disabled={isSubmitting} size="lg">
          {isSubmitting && <Loader2 className="size-4 animate-spin" />}
          {isSubmitting ? "Fetching feed..." : "Preview feed"}
        </Button>
      </form>
    </div>
  )
}

// --- Preview loading ---

function PreviewLoading() {
  return (
    <div className="flex flex-col items-center gap-4 py-16">
      <Loader2 className="size-6 animate-spin text-muted-foreground" />
      <p className="text-sm text-muted-foreground">
        Fetching and parsing feed...
      </p>
    </div>
  )
}

// --- Preview phase ---

function PreviewPhase({
  preview,
  categories,
  selectedCategoryId,
  onSelectCategory,
  onSubscribe,
  onBack,
}: {
  preview: PreviewFeedBody
  categories: Category[]
  selectedCategoryId: number | ""
  onSelectCategory: (id: number | "") => void
  onSubscribe: () => void
  onBack: () => void
}) {
  const items = preview.items ?? []

  return (
    <div className="flex flex-col gap-6">
      {/* Feed header */}
      <div className="flex items-start gap-4">
        {preview.favicon_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={preview.favicon_url}
            alt=""
            className="size-10 shrink-0 rounded-lg"
            width={40}
            height={40}
          />
        ) : (
          <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted">
            <Rss className="size-5 text-muted-foreground" />
          </div>
        )}
        <div className="min-w-0 flex-1">
          <h1 className="truncate font-serif text-xl font-bold tracking-tight">
            {preview.title || "Untitled feed"}
          </h1>
          {preview.site_url && (
            <a
              href={preview.site_url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              <span className="truncate">{preview.site_url}</span>
              <ExternalLink className="size-3 shrink-0" />
            </a>
          )}
          {preview.description && (
            <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
              {preview.description}
            </p>
          )}
        </div>
      </div>

      {/* Articles preview */}
      <div className="rounded-lg border border-border">
        <div className="border-b border-border px-4 py-2.5">
          <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Recent articles
            {items.length > 0 && (
              <span className="ml-1.5 text-muted-foreground/70">
                ({items.length})
              </span>
            )}
          </h2>
        </div>
        {items.length === 0 ? (
          <p className="px-4 py-6 text-sm text-muted-foreground">
            No articles found in this feed.
          </p>
        ) : (
          <ul className="divide-y divide-border">
            {items.map((item, idx) => (
              <li key={idx} className="px-4 py-3">
                <div className="flex flex-col gap-1">
                  <div className="flex items-baseline justify-between gap-3">
                    <a
                      href={item.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="line-clamp-1 text-sm font-medium hover:text-primary transition-colors"
                    >
                      {item.title || "Untitled"}
                    </a>
                    <time className="shrink-0 text-xs text-muted-foreground">
                      {formatDate(item.published_at)}
                    </time>
                  </div>
                  {item.author && (
                    <p className="text-xs text-muted-foreground">
                      {item.author}
                    </p>
                  )}
                  {item.description && (
                    <p className="line-clamp-2 text-sm text-muted-foreground">
                      {item.description}
                    </p>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Category selector + subscribe */}
      <div className="flex flex-col gap-4">
        {categories.length > 0 && (
          <div className="flex flex-col gap-2">
            <Label htmlFor="category">Category (optional)</Label>
            <select
              id="category"
              value={selectedCategoryId}
              onChange={(e) =>
                onSelectCategory(
                  e.target.value === "" ? "" : Number(e.target.value)
                )
              }
              className={cn(
                "h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm",
                "focus-visible:border-ring focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30",
                "transition-[color,box-shadow]"
              )}
            >
              <option value="">No category</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.id}>
                  {cat.title}
                </option>
              ))}
            </select>
          </div>
        )}

        <div className="flex items-center gap-2">
          <Button onClick={onSubscribe} size="lg">
            <Plus className="size-4" />
            Subscribe to this feed
          </Button>
          <Button variant="ghost" onClick={onBack} size="lg">
            Try another URL
          </Button>
        </div>
      </div>
    </div>
  )
}

// --- Subscribing phase ---

function SubscribingPhase({ preview }: { preview: PreviewFeedBody }) {
  return (
    <div className="flex flex-col items-center gap-4 py-16">
      <Loader2 className="size-6 animate-spin text-primary" />
      <p className="text-sm text-muted-foreground">
        Subscribing to &ldquo;{preview.title || preview.feed_url}&rdquo;...
      </p>
    </div>
  )
}

// --- Helpers ---

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return ""
  return date.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  })
}
