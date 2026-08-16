"use client";

import { useState, useCallback, type DragEvent } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ChevronRight, FolderOpen, Folder as FolderIcon } from "lucide-react";
import { getClient, updateFeed, updateFolder } from "@/lib/planetary";
import { getApiErrorMessage } from "@/lib/errors";
import { FeedIcon } from "@/components/feed-icon";
import { cn } from "@workspace/ui/lib/utils";
import type { Feed, Folder } from "@/lib/types";

type TreeNode =
  | { type: "folder"; folder: Folder; children: TreeNode[] }
  | { type: "feed"; feed: Feed };

type DragData = { kind: "feed"; id: number } | { kind: "folder"; id: number };

function isActive(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/";
  return pathname === href || pathname.startsWith(href + "/");
}

/**
 * Builds a tree from flat folder + feed lists. Folders are nested via parent_id.
 * Feeds are placed inside their folder_id, or at the root if unassigned.
 */
function buildTree(feeds: Feed[], folders: Folder[]): TreeNode[] {
  const folderMap = new Map<number, Folder>();
  for (const f of folders) folderMap.set(f.id, f);

  const childrenMap = new Map<number | null, TreeNode[]>();
  for (const f of folders) {
    const parentKey = f.parent_id ?? null;
    if (!childrenMap.has(parentKey)) childrenMap.set(parentKey, []);
    childrenMap.get(parentKey)!.push({
      type: "folder",
      folder: f,
      children: [],
    });
  }

  // Recursively populate folder children
  function populate(nodes: TreeNode[]) {
    for (const node of nodes) {
      if (node.type === "folder") {
        node.children = childrenMap.get(node.folder.id) ?? [];
        // Add feeds belonging to this folder
        const folderFeeds = feeds.filter((feed) => feed.folder_id === node.folder.id);
        node.children.push(...folderFeeds.map((feed) => ({ type: "feed", feed }) as TreeNode));
        populate(node.children);
      }
    }
  }

  const rootNodes = childrenMap.get(null) ?? [];
  // Root-level feeds (no folder)
  const rootFeeds = feeds.filter((feed) => !feed.folder_id);
  const rootTree = [...rootNodes, ...rootFeeds.map((feed) => ({ type: "feed", feed }) as TreeNode)];
  populate(rootNodes);
  return rootTree;
}

/** Check if a folder is a descendant of another folder (to prevent cycles). */
function isDescendant(folders: Folder[], folderId: number, ancestorId: number): boolean {
  const folderMap = new Map(folders.map((f) => [f.id, f]));
  let current = folderMap.get(folderId);
  while (current?.parent_id) {
    if (current.parent_id === ancestorId) return true;
    current = folderMap.get(current.parent_id);
  }
  return false;
}

function FeedNode({
  feed,
  onDragStart,
  onDragEnd,
  isDropTarget,
  onDrop,
}: {
  feed: Feed;
  onDragStart: (e: DragEvent, feedId: number) => void;
  onDragEnd: () => void;
  isDropTarget: boolean;
  onDrop: (e: DragEvent, targetFolderId: number | null) => void;
}) {
  const pathname = usePathname();
  const href = `/feeds/${feed.id}`;
  const active = isActive(pathname, href);

  return (
    <Link
      href={href}
      draggable
      onDragStart={(e) => onDragStart(e, feed.id)}
      onDragEnd={onDragEnd}
      onDragOver={(e) => {
        e.preventDefault();
        e.stopPropagation();
      }}
      onDrop={(e) => onDrop(e, null)}
      aria-current={active ? "page" : undefined}
      className={cn(
        "flex items-center gap-2.5 rounded-md px-3 py-2 text-sm",
        "text-sidebar-foreground/70 transition-colors cursor-pointer active:cursor-grabbing",
        active
          ? "bg-sidebar-accent text-sidebar-accent-foreground"
          : "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
        isDropTarget && "ring-1 ring-primary/50 ring-inset",
      )}
    >
      <FeedIcon siteUrl={feed.site_url} className="size-3.5 shrink-0 rounded-sm" />
      <span className="truncate">{feed.title}</span>
    </Link>
  );
}

function FolderNode({
  folder,
  childNodes,
  folders,
  feeds,
  onDragStart,
  onDragEnd,
  dragData,
  isDropTarget,
  dropTarget,
  setDropTarget,
  onDrop,
  depth,
}: {
  folder: Folder;
  childNodes: TreeNode[];
  folders: Folder[];
  feeds: Feed[];
  onDragStart: (e: DragEvent, data: DragData) => void;
  onDragEnd: () => void;
  dragData: DragData | null;
  isDropTarget: boolean;
  dropTarget: number | null;
  setDropTarget: (id: number | null) => void;
  onDrop: (e: DragEvent, targetFolderId: number | null) => void;
  depth: number;
}) {
  const pathname = usePathname();
  const [open, setOpen] = useState(true);

  const href = `/folders/${folder.id}`;
  const active = isActive(pathname, href);

  const canDropHere =
    dragData !== null &&
    !(
      dragData.kind === "folder" &&
      (dragData.id === folder.id || isDescendant(folders, folder.id, dragData.id))
    );

  return (
    <div>
      <div
        draggable
        onDragStart={(e) => onDragStart(e, { kind: "folder", id: folder.id })}
        onDragEnd={onDragEnd}
        onDragOver={(e) => {
          if (canDropHere) {
            e.preventDefault();
            e.stopPropagation();
            setDropTarget(folder.id);
          }
        }}
        onDrop={(e) => {
          if (canDropHere) {
            e.preventDefault();
            e.stopPropagation();
            onDrop(e, folder.id);
          }
        }}
        className={cn(
          "group flex items-center gap-2.5 rounded-md px-3 py-2 text-sm",
          "text-sidebar-foreground/70 transition-colors cursor-pointer active:cursor-grabbing",
          active
            ? "bg-sidebar-accent text-sidebar-accent-foreground"
            : "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
          isDropTarget && "ring-1 ring-primary/50 ring-inset",
        )}
        style={{ paddingLeft: `${depth * 16 + 12}px` }}
      >
        <Link
          href={href}
          aria-current={active ? "page" : undefined}
          className="flex min-w-0 flex-1 items-center gap-2.5"
        >
          {open ? (
            <FolderOpen
              className={cn("size-3.5 shrink-0", active ? "text-primary" : "text-muted-foreground")}
            />
          ) : (
            <FolderIcon
              className={cn("size-3.5 shrink-0", active ? "text-primary" : "text-muted-foreground")}
            />
          )}
          <span className="truncate">{folder.title}</span>
        </Link>
        <button
          type="button"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            setOpen((o) => !o);
          }}
          className="flex size-5 shrink-0 items-center justify-center text-muted-foreground hover:text-foreground"
          aria-label={open ? "Collapse" : "Expand"}
        >
          <ChevronRight
            className={cn("size-3.5 transition-transform duration-150", open && "rotate-90")}
          />
        </button>
      </div>
      {open && (
        <div className="flex flex-col gap-0.5">
          {childNodes.map((child) =>
            child.type === "folder" ? (
              <FolderNode
                key={`folder-${child.folder.id}`}
                folder={child.folder}
                childNodes={child.children}
                folders={folders}
                feeds={feeds}
                onDragStart={onDragStart}
                onDragEnd={onDragEnd}
                dragData={dragData}
                isDropTarget={dropTarget === child.folder.id}
                dropTarget={dropTarget}
                setDropTarget={setDropTarget}
                onDrop={onDrop}
                depth={depth + 1}
              />
            ) : (
              <div key={`feed-${child.feed.id}`} style={{ paddingLeft: `${(depth + 1) * 16}px` }}>
                <FeedNode
                  feed={child.feed}
                  onDragStart={(e, feedId) => onDragStart(e, { kind: "feed", id: feedId })}
                  onDragEnd={onDragEnd}
                  isDropTarget={false}
                  onDrop={onDrop}
                />
              </div>
            ),
          )}
        </div>
      )}
    </div>
  );
}

/**
 * FeedTree renders folders (nestable) and feeds in a single drag-and-drop tree.
 * Dragging a feed onto a folder moves it there. Dragging a folder onto another
 * folder nests it. Clicking a folder navigates to /folders/[id], clicking a
 * feed navigates to /feeds/[id].
 */
export function FeedTree({ feeds, folders }: { feeds: Feed[]; folders: Folder[] }) {
  const queryClient = useQueryClient();
  const [dragData, setDragData] = useState<DragData | null>(null);
  const [dropTarget, setDropTarget] = useState<number | null>(null);

  const tree = buildTree(feeds, folders);

  const handleDragStart = useCallback((e: DragEvent, data: DragData) => {
    setDragData(data);
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", JSON.stringify(data));
  }, []);

  const handleDragEnd = useCallback(() => {
    setDragData(null);
    setDropTarget(null);
  }, []);

  const handleDrop = useCallback(
    async (_e: DragEvent, targetFolderId: number | null) => {
      if (!dragData) return;
      const data = dragData;
      setDragData(null);
      setDropTarget(null);

      try {
        if (data.kind === "feed") {
          const { error } = await updateFeed({
            client: await getClient(),
            path: { feedId: data.id },
            body: { folder_id: targetFolderId ?? undefined },
          });
          if (error) throw error;
          await queryClient.invalidateQueries({ queryKey: ["feeds"] });
          await queryClient.invalidateQueries({ queryKey: ["entries"] });
        } else {
          // Moving a folder — need to update both title and parent_id
          const folder = folders.find((f) => f.id === data.id);
          if (!folder) return;
          if (targetFolderId === folder.parent_id) return; // no change
          const { error } = await updateFolder({
            client: await getClient(),
            path: { folderId: data.id },
            body: {
              title: folder.title,
              parent_id: targetFolderId ?? undefined,
            },
          });
          if (error) throw error;
          await queryClient.invalidateQueries({ queryKey: ["folders"] });
          await queryClient.invalidateQueries({ queryKey: ["feeds"] });
          await queryClient.invalidateQueries({ queryKey: ["entries"] });
        }
      } catch (err) {
        toast.error(getApiErrorMessage(err, "Could not move item"));
      }
    },
    [dragData, folders, queryClient],
  );

  return (
    <div className="flex flex-col gap-0.5">
      {tree.map((node) =>
        node.type === "folder" ? (
          <FolderNode
            key={`folder-${node.folder.id}`}
            folder={node.folder}
            childNodes={node.children}
            folders={folders}
            feeds={feeds}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            dragData={dragData}
            isDropTarget={dropTarget === node.folder.id}
            dropTarget={dropTarget}
            setDropTarget={setDropTarget}
            onDrop={handleDrop}
            depth={0}
          />
        ) : (
          <FeedNode
            key={`feed-${node.feed.id}`}
            feed={node.feed}
            onDragStart={(e, feedId) => handleDragStart(e, { kind: "feed", id: feedId })}
            onDragEnd={handleDragEnd}
            isDropTarget={false}
            onDrop={handleDrop}
          />
        ),
      )}
    </div>
  );
}
