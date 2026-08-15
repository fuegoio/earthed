import { PageHeader } from "@/components/page-header"
import { EntryTimeline } from "@/components/entry-timeline"

export const metadata = { title: "Unread" }

export default function UnreadPage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader title="Unread" />
      <EntryTimeline
        filter={{ status: "unread" }}
        emptyTitle="No unread articles"
        emptyDescription="You're all caught up. New articles will show up here as your feeds update."
      />
    </div>
  )
}
