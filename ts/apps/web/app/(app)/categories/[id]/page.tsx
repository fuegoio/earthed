import { notFound } from "next/navigation"
import { getClient, listCategories } from "@/lib/planetary"
import { getApiErrorMessage, apiErrorStatus } from "@/lib/errors"
import { ApiError } from "@/components/api-error"
import { EntryTimeline } from "@/components/entry-timeline"
import { Folder } from "lucide-react"
import type { Metadata } from "next"
import type { Category } from "@/lib/types"

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>
}): Promise<Metadata> {
  const { id } = await params
  const categoryId = Number(id)
  if (!Number.isFinite(categoryId)) return { title: "Category" }
  try {
    const { data } = await listCategories({ client: await getClient() })
    if (data) {
      const cats = data as Category[]
      const cat = cats.find((c) => c.id === categoryId)
      if (cat) return { title: cat.title }
    }
  } catch {
    // metadata is best-effort; fall through to default
  }
  return { title: "Category" }
}

export default async function CategoryPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const categoryId = Number(id)
  if (!Number.isFinite(categoryId)) notFound()

  const client = await getClient()
  const { data: categories, error } = await listCategories({ client })
  if (error) {
    return (
      <div className="p-4">
        <ApiError message={getApiErrorMessage(error)} status={apiErrorStatus(error)} />
      </div>
    )
  }

  const category = (categories as Category[] | undefined)?.find(
    (c) => c.id === categoryId
  )
  if (!category) notFound()

  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <h1 className="flex items-center gap-2 font-serif text-lg font-bold tracking-tight">
          <Folder className="size-4 text-muted-foreground" />
          {category.title}
        </h1>
      </div>
      <EntryTimeline
        filter={{ category_id: category.id }}
        emptyTitle="No articles in this category"
        emptyDescription="Feeds in this category haven't produced any entries yet."
      />
    </div>
  )
}
