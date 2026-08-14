import { notFound } from "next/navigation"
import { getClient, listCategories } from "@/lib/planetary"
import { getApiErrorMessage, apiErrorStatus } from "@/lib/errors"
import { ApiError } from "@/components/api-error"
import { EntryTimeline } from "@/components/entry-timeline"
import { Folder } from "lucide-react"
import type { Category } from "@/lib/types"

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
