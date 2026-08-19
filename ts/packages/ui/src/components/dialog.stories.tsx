import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Pencil, Plus } from "lucide-react";

import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@workspace/ui/components/dialog";
import { Button } from "@workspace/ui/components/button";
import { Input } from "@workspace/ui/components/input";
import { Label } from "@workspace/ui/components/label";

const meta: Meta<typeof Dialog> = {
  title: "UI/Dialog",
  component: Dialog,
  tags: ["autodocs"],
  parameters: {
    layout: "centered",
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

/** Basic dialog with a title, description, and confirm/cancel footer. */
export const Default: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger render={<Button>Open dialog</Button>} />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Are you sure?</DialogTitle>
          <DialogDescription>
            This action cannot be undone. All associated data will be permanently removed.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose render={<Button variant="ghost">Cancel</Button>} />
          <Button variant="destructive">Confirm</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

/** Dialog containing a form — the common case for create/edit flows. */
export const WithForm: Story = {
  render: () => {
    function FormDialog() {
      const [value, setValue] = useState("");
      return (
        <Dialog>
          <DialogTrigger
            render={
              <Button>
                <Plus />
                New folder
              </Button>
            }
          />
          <DialogContent>
            <DialogHeader>
              <DialogTitle>New folder</DialogTitle>
              <DialogDescription>Group related feeds together.</DialogDescription>
            </DialogHeader>
            <form
              className="mt-4 flex flex-col gap-3"
              onSubmit={(e) => e.preventDefault()}
            >
              <div className="flex flex-col gap-2">
                <Label htmlFor="story-folder-name">Title</Label>
                <Input
                  id="story-folder-name"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  placeholder="e.g. Tech, News, Design"
                  autoFocus
                />
              </div>
              <DialogFooter>
                <DialogClose render={<Button variant="ghost" type="button">Cancel</Button>} />
                <Button type="submit" disabled={!value.trim()}>
                  Create
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      );
    }
    return <FormDialog />;
  },
};

/** Icon-button trigger — used in toolbars and page headers. */
export const IconTrigger: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger
        render={
          <Button variant="ghost" size="icon-sm" aria-label="Rename">
            <Pencil className="size-3.5" />
          </Button>
        }
      />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Rename feed</DialogTitle>
          <DialogDescription>Give this feed a custom display name.</DialogDescription>
        </DialogHeader>
        <div className="mt-4 flex flex-col gap-2">
          <Label htmlFor="story-feed-name">Name</Label>
          <Input id="story-feed-name" defaultValue="Hacker News" autoFocus />
        </div>
        <DialogFooter className="mt-3">
          <DialogClose render={<Button variant="ghost">Cancel</Button>} />
          <Button>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

/** Wider dialog — pass a className override to DialogContent. */
export const Wide: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger render={<Button>Open wide dialog</Button>} />
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>New API token</DialogTitle>
          <DialogDescription>Give it a label so you remember where it&apos;s used.</DialogDescription>
        </DialogHeader>
        <div className="mt-4 flex flex-col gap-2">
          <Label htmlFor="story-token-label">Label</Label>
          <Input id="story-token-label" placeholder="e.g. earthed-cli" autoFocus />
        </div>
        <DialogFooter className="mt-3">
          <DialogClose render={<Button variant="ghost">Cancel</Button>} />
          <Button>Create token</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};
