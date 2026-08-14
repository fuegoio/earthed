import { notFound } from "next/navigation"
import { getClient, getEntry, getFeed } from "@/lib/planetary"
import { getApiErrorMessage, apiErrorStatus } from "@/lib/errors"
import { ApiError } from "@/components/api-error"
import { EntryReader } from "@/components/entry-reader"
import type { Entry, Feed } from "@/lib/types"

export default async function EntryPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const entryId = Number(id)
  if (!Number.isFinite(entryId)) notFound()

  const client = await getClient()

  const { data: entry, error } = await getEntry({
    client,
    path: { entryId },
  })
  if (error) {
    if (apiErrorStatus(error) === 404) notFound()
    return (
      <div className="p-4">
        <ApiError message={getApiErrorMessage(error)} status={apiErrorStatus(error)} />
      </div>
    )
  }
  if (!entry) notFound()

  let feed: Feed | undefined
  if (entry.feed_id) {
    const { data: f } = await getFeed({
      client,
      path: { feedId: entry.feed_id },
    })
    feed = (f as Feed | undefined) ?? undefined
  }

  return <EntryReader entry={entry as Entry} feed={feed} />
}
