import { Search } from "lucide-react"
import { EntryTimeline } from "@/components/entry-timeline"
import { PageHeader } from "@/components/page-header"
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"

export const metadata = { title: "Search" }

export default async function SearchPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>
}) {
  const { q } = await searchParams
  const query = (q ?? "").trim()

  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader
        title={query ? `Results for “${query}”` : "Search"}
        icon={<Search className="size-4 text-muted-foreground" />}
      />
      {query ? (
        <EntryTimeline
          filter={{ search: query }}
          emptyTitle="No matching articles"
          emptyDescription="No entries matched your search. Try different keywords."
        />
      ) : (
        <div className="p-4">
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
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
