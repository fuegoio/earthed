import { Search } from "lucide-react"
import { EntryTimeline } from "@/components/entry-timeline"
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/empty"

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
        <div className="p-4">
          <Empty>
            <EmptyHeader>
              <EmptyMedia>
                <Search className="size-6 text-primary" />
              </EmptyMedia>
              <EmptyTitle>Search your feeds</EmptyTitle>
              <EmptyDescription>
                Type in the search box above to find articles across all your
                subscriptions. Search matches titles and article content.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        </div>
      )}
    </div>
  )
}
