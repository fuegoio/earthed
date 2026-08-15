import { PageHeader } from "@/components/page-header"
import { Logo } from "@/components/logo"
import { EntryTimeline } from "@/components/entry-timeline"

export const metadata = { title: "Timeline" }

export default function HomePage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader title="Timeline" icon={<Logo className="size-4 text-primary" />} />
      <EntryTimeline
        filter={{}}
        emptyTitle="Your timeline is empty"
        emptyDescription="Subscribe to RSS feeds and your latest articles will appear here."
      />
    </div>
  )
}
