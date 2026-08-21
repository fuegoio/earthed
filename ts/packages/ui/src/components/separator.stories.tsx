import type { Meta, StoryObj } from "@storybook/react";

import { Separator } from "@workspace/ui/components/separator";

const meta: Meta<typeof Separator> = {
  title: "UI/Separator",
  component: Separator,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="flex w-full max-w-sm flex-col gap-4 rounded-lg bg-sidebar p-4">
        <span className="text-sm text-sidebar-foreground">Section above</span>
        <Story />
        <span className="text-sm text-sidebar-foreground">Section below</span>
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Multiple: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <span className="text-sm text-sidebar-foreground">Feeds</span>
      <Separator />
      <span className="text-sm text-sidebar-foreground">Lists</span>
      <Separator />
      <span className="text-sm text-sidebar-foreground">Settings</span>
    </div>
  ),
};
