import type { Meta, StoryObj } from "@storybook/react";
import { RssIcon } from "lucide-react";

import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
  EmptyDescription,
  EmptyContent,
} from "@workspace/ui/components/empty";
import { Button } from "@workspace/ui/components/button";

const meta: Meta<typeof Empty> = {
  title: "UI/Empty",
  component: Empty,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="flex min-h-[400px] max-w-2xl items-center justify-center">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Empty>
      <EmptyHeader>
        <EmptyTitle>No feeds yet</EmptyTitle>
        <EmptyDescription>
          Subscribe to your favorite sites and their latest entries will appear here.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button>Add your first feed</Button>
      </EmptyContent>
    </Empty>
  ),
};

export const WithIcon: Story = {
  render: () => (
    <Empty>
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <RssIcon />
        </EmptyMedia>
        <EmptyTitle>No entries</EmptyTitle>
        <EmptyDescription>
          This feed has no entries yet. Check back later or <a href="#">add another feed</a>.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button variant="outline">Refresh</Button>
      </EmptyContent>
    </Empty>
  ),
};

export const MediaDefault: Story = {
  render: () => (
    <Empty>
      <EmptyHeader>
        <EmptyMedia variant="default">
          <RssIcon className="size-10" />
        </EmptyMedia>
        <EmptyTitle>Large icon</EmptyTitle>
        <EmptyDescription>The default media variant is transparent.</EmptyDescription>
      </EmptyHeader>
    </Empty>
  ),
};
