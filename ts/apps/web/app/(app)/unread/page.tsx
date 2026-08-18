import { PageHeader } from "@/components/page-header";
import { EntryTimeline } from "@/components/entry-timeline";
import { MarkAllReadButton } from "@/components/mark-all-read-button";

export const metadata = { title: "Unread" };

export default function UnreadPage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader title="Unread" actions={<MarkAllReadButton />} />
      <EntryTimeline
        filter={{ status: "unread" }}
        emptyTitle="You're all caught up"
        emptyDescription="Nothing left to read. Enjoy the quiet — new articles will land here when your feeds update."
        emptyVariant="celebration"
      />
    </div>
  );
}
