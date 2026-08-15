import { Circle } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { EntryTimeline } from "@/components/entry-timeline"

export const metadata = { title: "Unread" }

export default function UnreadPage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader
        title="Unread"
        icon={<Circle className="size-4 fill-primary text-primary" />}
      />
      <EntryTimeline
        filter={{ status: "unread" }}
        emptyTitle="No unread articles"
        emptyDescription="You're all caught up. New articles will show up here as your feeds update."
      />
    </div>
  )
}
