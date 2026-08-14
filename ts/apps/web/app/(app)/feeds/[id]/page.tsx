import { notFound } from "next/navigation"
import { getClient, getFeed } from "@/lib/planetary"
import { getApiErrorMessage, apiErrorStatus } from "@/lib/errors"
import { ApiError } from "@/components/api-error"
import { FeedDetail } from "@/components/feed-detail"
import type { Feed } from "@/lib/types"

export default async function FeedPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const feedId = Number(id)
  if (!Number.isFinite(feedId)) notFound()

  const client = await getClient()
  const { data: feed, error } = await getFeed({
    client,
    path: { feedId },
  })
  if (error) {
    if (apiErrorStatus(error) === 404) notFound()
    return (
      <div className="p-4">
        <ApiError message={getApiErrorMessage(error)} status={apiErrorStatus(error)} />
      </div>
    )
  }
  if (!feed) notFound()

  return <FeedDetail feed={feed as Feed} />
}
