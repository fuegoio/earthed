import type { Meta, StoryObj } from "@storybook/react";

import { Label } from "@workspace/ui/components/label";
import { Input } from "@workspace/ui/components/input";

const meta: Meta<typeof Label> = {
  title: "UI/Label",
  component: Label,
  tags: ["autodocs"],
  args: { children: "Email address" },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  decorators: [(Story) => <Story />],
};

export const WithInput: Story = {
  render: () => (
    <div className="flex max-w-sm flex-col gap-2">
      <Label htmlFor="email">Email address</Label>
      <Input id="email" type="email" placeholder="you@example.com" />
    </div>
  ),
};

export const WithCheckbox: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <input id="terms" type="checkbox" className="size-4" />
      <Label htmlFor="terms">I agree to the terms</Label>
    </div>
  ),
};

export const DisabledPeer: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <input id="disabled-cb" type="checkbox" disabled className="size-4" />
      <Label htmlFor="disabled-cb">This label dims with its disabled peer</Label>
    </div>
  ),
};
