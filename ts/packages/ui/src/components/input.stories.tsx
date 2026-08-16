import type { Meta, StoryObj } from "@storybook/react";

import { Input } from "@workspace/ui/components/input";

const meta: Meta<typeof Input> = {
  title: "UI/Input",
  component: Input,
  tags: ["autodocs"],
  args: {
    placeholder: "Enter feed URL",
  },
  argTypes: {
    type: {
      control: "select",
      options: ["text", "email", "password", "search", "url", "number"],
    },
  },
  decorators: [(Story) => <div className="max-w-sm"><Story /></div>],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithLabel: Story = {
  render: () => (
    <div className="flex flex-col gap-2">
      <label htmlFor="feed-url" className="text-sm font-medium">Feed URL</label>
      <Input id="feed-url" type="url" placeholder="https://example.com/feed.xml" />
    </div>
  ),
};

export const Password: Story = {
  args: { type: "password", placeholder: "••••••••" },
};

export const Disabled: Story = {
  args: { disabled: true, value: "Cannot edit this" },
};

export const Invalid: Story = {
  args: { "aria-invalid": true, defaultValue: "not-a-url" },
};

export const FileInput: Story = {
  args: { type: "file" },
};
