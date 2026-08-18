import { PageHeader } from "@/components/page-header";
import { EntryTimeline } from "@/components/entry-timeline";

export const metadata = { title: "All" };

export default function AllPage() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader title="All" />
      <EntryTimeline
        filter={{}}
        emptyTitle="Your timeline is empty"
        emptyDescription="Subscribe to RSS feeds and your latest articles will appear here."
      />
    </div>
  );
}
