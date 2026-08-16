import type { Meta, StoryObj } from "@storybook/react";

import { Textarea } from "@workspace/ui/components/textarea";

const meta: Meta<typeof Textarea> = {
  title: "UI/Textarea",
  component: Textarea,
  tags: ["autodocs"],
  args: { placeholder: "Write something…" },
  decorators: [(Story) => <div className="max-w-sm"><Story /></div>],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithLabel: Story = {
  render: () => (
    <div className="flex flex-col gap-2">
      <label htmlFor="bio" className="text-sm font-medium">Bio</label>
      <Textarea id="bio" placeholder="Tell us about yourself" rows={4} />
    </div>
  ),
};

export const Disabled: Story = {
  args: { disabled: true, defaultValue: "This field is locked." },
};

export const Invalid: Story = {
  args: { "aria-invalid": true, defaultValue: "Too short" },
};

export const Prefilled: Story = {
  args: { defaultValue: "https://example.com/feed.xml\n\nThis is a multi-line\nfield." },
};
