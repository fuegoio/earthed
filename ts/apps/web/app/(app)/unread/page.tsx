import { Circle } from "lucide-react"
import { EntryTimeline } from "@/components/entry-timeline"

export const metadata = { title: "Unread" }

export default function UnreadPage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <h1 className="flex items-center gap-2 font-serif text-lg font-bold tracking-tight">
          <Circle className="size-4 fill-primary text-primary" />
          Unread
        </h1>
      </div>
      <EntryTimeline
        filter={{ status: "unread" }}
        emptyTitle="No unread articles"
        emptyDescription="You're all caught up. New articles will show up here as your feeds update."
      />
    </div>
  )
}
