import { EntryTimeline } from "@/components/entry-timeline"

export default function HomePage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <h1 className="font-serif text-lg font-bold tracking-tight">
          Timeline
        </h1>
      </div>
      <EntryTimeline
        filter={{}}
        emptyTitle="Your timeline is empty"
        emptyDescription="Subscribe to RSS feeds and your latest articles will appear here."
      />
    </div>
  )
}
