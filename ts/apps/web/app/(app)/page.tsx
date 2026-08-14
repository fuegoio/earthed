import Link from "next/link"
import { Rss, Plus } from "lucide-react"
import { buttonVariants } from "@workspace/ui/components/button"
import { cn } from "@workspace/ui/lib/utils"

export default function HomePage() {
  return (
    <div className="flex flex-1 items-center justify-center p-8">
      <div className="flex max-w-md flex-col items-center gap-5 text-center">
        <div className="flex size-16 items-center justify-center rounded-2xl bg-primary/10">
          <Rss className="size-8 text-primary" />
        </div>
        <div className="flex flex-col gap-2">
          <h2 className="font-serif text-xl font-bold tracking-tight">
            Your timeline will appear here
          </h2>
          <p className="text-sm text-muted-foreground">
            Subscribe to RSS feeds to start reading. Your latest entries will
            show up in this timeline.
          </p>
        </div>
        <Link
          href="/feeds/new"
          className={cn(buttonVariants({ size: "lg" }))}
        >
          <Plus className="size-4" />
          Subscribe to a feed
        </Link>
      </div>
    </div>
  )
}
