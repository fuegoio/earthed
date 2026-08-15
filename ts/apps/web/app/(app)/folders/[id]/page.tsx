import { notFound } from "next/navigation"
import { getClient, listFolders } from "@/lib/planetary"
import { getApiErrorMessage, apiErrorStatus } from "@/lib/errors"
import { ApiError } from "@/components/api-error"
import { EntryTimeline } from "@/components/entry-timeline"
import { PageHeader } from "@/components/page-header"
import { FolderDeleteButton } from "@/components/folder-delete-button"
import { FolderIcon } from "lucide-react"
import type { Metadata } from "next"
import type { Folder } from "@/lib/types"

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>
}): Promise<Metadata> {
  const { id } = await params
  const folderId = Number(id)
  if (!Number.isFinite(folderId)) return { title: "Folder" }
  try {
    const { data } = await listFolders({ client: await getClient() })
    if (data) {
      const folders = data as Folder[]
      const folder = folders.find((f) => f.id === folderId)
      if (folder) return { title: folder.title }
    }
  } catch {
    // metadata is best-effort; fall through to default
  }
  return { title: "Folder" }
}

export default async function FolderPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const folderId = Number(id)
  if (!Number.isFinite(folderId)) notFound()

  const client = await getClient()
  const { data: folders, error } = await listFolders({ client })
  if (error) {
    return (
      <div className="p-4">
        <ApiError message={getApiErrorMessage(error)} status={apiErrorStatus(error)} />
      </div>
    )
  }

  const folder = (folders as Folder[] | undefined)?.find(
    (f) => f.id === folderId
  )
  if (!folder) notFound()

  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader
        title={folder.title}
        icon={<FolderIcon className="size-4 text-muted-foreground" />}
        actions={<FolderDeleteButton folder={folder} />}
      />
      <EntryTimeline
        filter={{ folder_id: folder.id }}
        emptyTitle="No articles in this folder"
        emptyDescription="Feeds in this folder haven't produced any entries yet."
      />
    </div>
  )
}
