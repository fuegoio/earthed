import { Rss } from "lucide-react"
import { faviconUrl } from "@/lib/format"

/** A feed favicon, falling back to the RSS icon when no site URL is known. */
export function FeedIcon({
  siteUrl,
  className,
}: {
  siteUrl?: string
  className?: string
}) {
  if (siteUrl) {
    return (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={faviconUrl(siteUrl)}
        alt=""
        width={16}
        height={16}
        className={className ?? "size-4 shrink-0 rounded-sm"}
      />
    )
  }
  return <Rss className={className ?? "size-4 shrink-0 text-muted-foreground"} />
}
