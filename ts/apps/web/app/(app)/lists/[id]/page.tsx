import { notFound } from "next/navigation"
import { getClient, getFeedList, getMe } from "@/lib/planetary"
import { getApiErrorMessage, apiErrorStatus } from "@/lib/errors"
import { ApiError } from "@/components/api-error"
import { FeedListDetail } from "@/components/feed-list-detail"
import type { FeedList, User } from "@/lib/types"

export default async function FeedListPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const listId = Number(id)
  if (!Number.isFinite(listId)) notFound()

  const client = await getClient()

  // Fetch the current user (for ownership check) and the list in parallel.
  const [{ data: me, error: meError }, { data: list, error }] = await Promise.all([
    getMe({ client }),
    getFeedList({ client, path: { listId } }),
  ])

  if (meError || !me) {
    notFound()
  }
  if (error) {
    if (apiErrorStatus(error) === 404) notFound()
    return (
      <div className="p-4">
        <ApiError message={getApiErrorMessage(error)} status={apiErrorStatus(error)} />
      </div>
    )
  }
  if (!list) notFound()

  return <FeedListDetail list={list as FeedList} currentUserId={(me as User).id} />
}
