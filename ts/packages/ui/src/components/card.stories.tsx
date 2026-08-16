import type { Meta, StoryObj } from "@storybook/react";
import { Bell, Heart } from "lucide-react";

import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@workspace/ui/components/card";
import { Button } from "@workspace/ui/components/button";

const meta: Meta<typeof Card> = {
  title: "UI/Card",
  component: Card,
  tags: ["autodocs"],
  decorators: [(Story) => <div className="max-w-sm"><Story /></div>],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Card>
      <CardHeader>
        <CardTitle>Notifications</CardTitle>
        <CardDescription>You have 3 unread messages.</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">
          Card content goes here. This is where the main body of the card lives.
        </p>
      </CardContent>
      <CardFooter className="gap-2">
        <Button size="sm">Mark all read</Button>
        <Button size="sm" variant="ghost">Dismiss</Button>
      </CardFooter>
    </Card>
  ),
};

export const WithoutFooter: Story = {
  render: () => (
    <Card>
      <CardHeader>
        <CardTitle>Feed settings</CardTitle>
        <CardDescription>Manage how this feed updates.</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">No footer needed here.</p>
      </CardContent>
    </Card>
  ),
};

export const WithoutContent: Story = {
  render: () => (
    <Card>
      <CardHeader>
        <CardTitle>Quick stats</CardTitle>
        <CardDescription>12 feeds · 847 unread</CardDescription>
      </CardHeader>
      <CardFooter>
        <Button size="sm" variant="ghost"><Bell className="size-4" /> Subscribe</Button>
      </CardFooter>
    </Card>
  ),
};

export const WithIcon: Story = {
  render: () => (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Heart className="size-4" /> Favorites
        </CardTitle>
        <CardDescription>Your starred entries across all feeds.</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">Nothing here yet.</p>
      </CardContent>
    </Card>
  ),
};
