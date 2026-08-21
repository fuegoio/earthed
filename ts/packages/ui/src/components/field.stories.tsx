import type { Meta, StoryObj } from "@storybook/react";

import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSeparator,
  FieldSet,
} from "@workspace/ui/components/field";
import { Input } from "@workspace/ui/components/input";
import { Button } from "@workspace/ui/components/button";

const meta: Meta<typeof Field> = {
  title: "UI/Field",
  component: Field,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="max-w-sm">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Field>
      <FieldLabel htmlFor="handle">Handle</FieldLabel>
      <Input id="handle" placeholder="alice.bsky.social" />
    </Field>
  ),
};

export const WithDescription: Story = {
  render: () => (
    <Field>
      <FieldLabel htmlFor="handle">Handle</FieldLabel>
      <Input id="handle" placeholder="alice.bsky.social" />
      <FieldDescription>
        The same account you use in the Bluesky app.
      </FieldDescription>
    </Field>
  ),
};

export const WithError: Story = {
  render: () => (
    <Field data-invalid>
      <FieldLabel htmlFor="email">Email</FieldLabel>
      <Input id="email" type="email" aria-invalid placeholder="you@example.com" />
      <FieldError>Enter a valid email address.</FieldError>
    </Field>
  ),
};

export const Horizontal: Story = {
  render: () => (
    <Field orientation="horizontal">
      <input id="remember" type="checkbox" className="size-4" />
      <FieldContent>
        <FieldLabel htmlFor="remember">Remember me</FieldLabel>
        <FieldDescription>Keep me logged in on this device.</FieldDescription>
      </FieldContent>
    </Field>
  ),
};

export const Group: Story = {
  render: () => (
    <FieldGroup>
      <Field>
        <FieldLabel htmlFor="name">Name</FieldLabel>
        <Input id="name" placeholder="Evil Rabbit" />
        <FieldDescription>This appears on your profile.</FieldDescription>
      </Field>
      <FieldSeparator />
      <Field>
        <FieldLabel htmlFor="username">Username</FieldLabel>
        <Input id="username" aria-invalid />
        <FieldError>Choose another username.</FieldError>
      </Field>
    </FieldGroup>
  ),
};

export const FieldsetExample: Story = {
  render: () => (
    <FieldSet>
      <FieldLegend>Profile</FieldLegend>
      <FieldDescription>This appears on invoices and emails.</FieldDescription>
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="full-name">Full name</FieldLabel>
          <Input id="full-name" placeholder="Evil Rabbit" />
        </Field>
        <Field>
          <FieldLabel htmlFor="username-2">Username</FieldLabel>
          <Input id="username-2" placeholder="evilrabbit" />
          <FieldDescription>Pick something memorable.</FieldDescription>
        </Field>
        <Button size="lg" className="w-full">
          Save
        </Button>
      </FieldGroup>
    </FieldSet>
  ),
};
