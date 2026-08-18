import { PageHeader } from "@/components/page-header";
import { EntryTimeline } from "@/components/entry-timeline";
import { MarkAllReadButton } from "@/components/mark-all-read-button";
import { ScrollArea } from "@workspace/ui/components/scroll-area";

export const metadata = { title: "Unread" };

export default function UnreadPage() {
  return (
    <div className="flex h-full flex-col overflow-hidden">
      <div className="mx-auto w-full max-w-3xl shrink-0">
        <PageHeader title="Unread" actions={<MarkAllReadButton />} />
      </div>
      <ScrollArea className="flex-1 min-h-0">
        <div className="mx-auto w-full max-w-3xl">
          <EntryTimeline
            filter={{ status: "unread" }}
            emptyTitle="You're all caught up"
            emptyDescription="Nothing left to read. Enjoy the quiet — new articles will land here when your feeds update."
            emptyVariant="celebration"
          />
        </div>
      </ScrollArea>
    </div>
  );
}
