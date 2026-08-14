import { Star } from "lucide-react"
import { EntryTimeline } from "@/components/entry-timeline"

export const metadata = { title: "Starred" }

export default function StarredPage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <h1 className="flex items-center gap-2 font-serif text-lg font-bold tracking-tight">
          <Star className="size-4 fill-amber-400 text-amber-400" />
          Starred
        </h1>
      </div>
      <EntryTimeline
        filter={{ starred: true }}
        emptyTitle="No starred articles"
        emptyDescription="Star articles you want to keep — they'll show up here for quick access."
      />
    </div>
  )
}
