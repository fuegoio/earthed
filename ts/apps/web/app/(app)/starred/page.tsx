import { Star } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { EntryTimeline } from "@/components/entry-timeline"

export const metadata = { title: "Starred" }

export default function StarredPage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader
        title="Starred"
        icon={<Star className="size-4 fill-primary text-primary" />}
      />
      <EntryTimeline
        filter={{ starred: true }}
        emptyTitle="No starred articles"
        emptyDescription="Star articles you want to keep — they'll show up here for quick access."
      />
    </div>
  )
}
