import { Search } from "lucide-react"
import { EntryTimeline } from "@/components/entry-timeline"

export default async function SearchPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>
}) {
  const { q } = await searchParams
  const query = (q ?? "").trim()

  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="flex items-center gap-2 border-b border-border px-4 py-3">
        <Search className="size-4 text-muted-foreground" />
        <h1 className="font-serif text-lg font-bold tracking-tight">
          {query ? `Results for “${query}”` : "Search"}
        </h1>
      </div>
      {query ? (
        <EntryTimeline
          filter={{ search: query }}
          emptyTitle="No matching articles"
          emptyDescription="No entries matched your search. Try different keywords."
        />
      ) : (
        <div className="px-4 py-16 text-center text-sm text-muted-foreground">
          Type in the search box above to find articles across all your feeds.
        </div>
      )}
    </div>
  )
}
