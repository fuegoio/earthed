import { Rss } from "lucide-react"

export default function HomePage() {
  return (
    <div className="flex flex-1 items-center justify-center p-8">
      <div className="flex max-w-md flex-col items-center gap-4 text-center">
        <div className="flex size-16 items-center justify-center rounded-2xl bg-muted">
          <Rss className="size-8 text-muted-foreground" />
        </div>
        <h2 className="text-xl font-semibold">Your timeline will appear here</h2>
        <p className="text-sm text-muted-foreground">
          Subscribe to RSS feeds to start reading. Your latest entries will
          show up in this timeline.
        </p>
      </div>
    </div>
  )
}
