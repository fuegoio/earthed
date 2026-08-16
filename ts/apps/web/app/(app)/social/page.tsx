"use client";

import { useEffect, useRef } from "react";
import Link from "next/link";
import { useInfiniteQuery, useQuery, type InfiniteData } from "@tanstack/react-query";
import { Users } from "lucide-react";
import { SharedArticleCard } from "@/components/shared-article-card";
import { EntryCardSkeleton } from "@/components/entry-card-skeleton";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty";
import { PageHeader } from "@/components/page-header";
import { getClient, socialTimeline, listFollowing, unwrap } from "@/lib/planetary";
import type { SharedArticle, UserProfile } from "@/lib/types";

const PAGE_SIZE = 50;

export default function SocialPage() {
  const { data: following } = useQuery<UserProfile[]>({
    queryKey: ["following"],
    queryFn: async () => unwrap(listFollowing({ client: await getClient() })),
  });

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading, error } =
    useInfiniteQuery<
      SharedArticle[],
      Error,
      InfiniteData<SharedArticle[]>,
      ["social-timeline"],
      number
    >({
      queryKey: ["social-timeline"],
      queryFn: async ({ pageParam }) => {
        const result = await socialTimeline({
          client: await getClient(),
          query: { limit: PAGE_SIZE, offset: pageParam },
        });
        if (result.error) throw result.error;
        return (result.data ?? []) as SharedArticle[];
      },
      initialPageParam: 0,
      getNextPageParam: (lastPage, _all, lastParam) =>
        lastPage.length < PAGE_SIZE ? undefined : lastParam + PAGE_SIZE,
    });

  const sentinelRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || !hasNextPage || isFetchingNextPage) return;
    const obs = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) fetchNextPage();
      },
      { rootMargin: "600px" },
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const articles = data?.pages.flat() ?? [];

  const followingCount = following?.length ?? 0;

  return (
    <div className="flex flex-col">
      <PageHeader title="Social" />

      {isLoading ? (
        <div className="divide-y divide-border">
          {Array.from({ length: 8 }).map((_, i) => (
            <EntryCardSkeleton key={i} />
          ))}
        </div>
      ) : followingCount === 0 ? (
        <div className="p-4">
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Users className="size-6 text-primary" />
              </EmptyMedia>
              <EmptyTitle>Follow people to see their shares</EmptyTitle>
              <EmptyDescription>
                When people you follow share articles, they will appear here.
                Visit a user profile to follow them.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        </div>
      ) : articles.length === 0 ? (
        <div className="p-4">
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Users className="size-6 text-primary" />
              </EmptyMedia>
              <EmptyTitle>No shares yet</EmptyTitle>
              <EmptyDescription>
                You follow {followingCount} {followingCount === 1 ? "person" : "people"}, but
                nobody has shared anything yet.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        </div>
      ) : (
        <div className="divide-y divide-border">
          {articles.map((article, i) => (
            <SharedArticleCard key={article.id} article={article} staggerIndex={i} />
          ))}
          <div ref={sentinelRef} className="h-px" />
          {isFetchingNextPage && (
            <div className="divide-y divide-border">
              {Array.from({ length: 3 }).map((_, i) => (
                <EntryCardSkeleton key={i} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
