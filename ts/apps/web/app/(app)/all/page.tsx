import { PageHeader } from "@/components/page-header";
import { EntryTimeline } from "@/components/entry-timeline";
import { ScrollArea } from "@workspace/ui/components/scroll-area";

export const metadata = { title: "All" };

export default function AllPage() {
  return (
    <div className="flex h-full flex-col overflow-hidden">
      <div className="mx-auto w-full max-w-3xl shrink-0">
        <PageHeader title="All" />
      </div>
      <ScrollArea className="flex-1 min-h-0">
        <div className="mx-auto w-full max-w-3xl">
          <EntryTimeline
            filter={{}}
            emptyTitle="Your timeline is empty"
            emptyDescription="Subscribe to RSS feeds and your latest articles will appear here."
          />
        </div>
      </ScrollArea>
    </div>
  );
}
