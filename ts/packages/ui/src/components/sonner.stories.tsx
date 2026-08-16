import type { Meta, StoryObj } from "@storybook/react";
import { toast } from "sonner";

import { Toaster } from "@workspace/ui/components/sonner";
import { Button } from "@workspace/ui/components/button";

const meta: Meta<typeof Toaster> = {
  title: "UI/Toaster",
  component: Toaster,
  tags: ["autodocs"],
  decorators: [(Story) => (
    <>
      <Story />
      <div className="flex flex-wrap gap-3">
        <Button size="sm" onClick={() => toast.success("Feed added")}>Success</Button>
        <Button size="sm" variant="outline" onClick={() => toast.info("Checking for updates")}>Info</Button>
        <Button size="sm" variant="outline" onClick={() => toast.warning("Rate limit approaching")}>Warning</Button>
        <Button size="sm" variant="destructive" onClick={() => toast.error("Failed to fetch feed")}>Error</Button>
        <Button size="sm" variant="ghost" onClick={() => toast.loading("Refreshing…")}>Loading</Button>
      </div>
    </>
  )],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithDescription: Story = {
  decorators: [(Story) => (
    <>
      <Story />
      <div className="flex flex-wrap gap-3">
        <Button size="sm" onClick={() => toast.success("Subscribed", { description: "New entries from this feed will appear in your timeline." })}>
          Subscribe
        </Button>
        <Button size="sm" variant="destructive" onClick={() => toast.error("Connection lost", { description: "Retrying in 30 seconds." })}>
          Error
        </Button>
      </div>
    </>
  )],
  render: () => <Toaster />,
};
