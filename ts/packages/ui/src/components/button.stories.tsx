import type { Meta, StoryObj } from "@storybook/react";
import { Plus, ChevronRight } from "lucide-react";

import { Button } from "@workspace/ui/components/button";

const meta = {
  title: "UI/Button",
  component: Button,
  tags: ["autodocs"],
  args: {
    children: "Subscribe",
  },
  argTypes: {
    variant: {
      control: "select",
      options: ["default", "outline", "secondary", "ghost", "destructive", "link"],
    },
    size: {
      control: "select",
      options: ["default", "xs", "sm", "lg", "icon", "icon-xs", "icon-sm", "icon-lg"],
    },
  },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: { variant: "default", size: "default" },
};

export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-4">
      <Button variant="default">Default</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="destructive">Destructive</Button>
      <Button variant="link">Link</Button>
    </div>
  ),
};

export const Sizes: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-4">
      <Button size="xs">Extra small</Button>
      <Button size="sm">Small</Button>
      <Button size="default">Default</Button>
      <Button size="lg">Large</Button>
      <Button size="icon-sm" aria-label="Add"><Plus /></Button>
      <Button size="icon" aria-label="Add"><Plus /></Button>
      <Button size="icon-lg" aria-label="Add"><Plus /></Button>
    </div>
  ),
};

export const WithIcon: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-4">
      <Button variant="default">
        <Plus data-icon="inline-start" /> New feed
      </Button>
      <Button variant="default">
        Continue <ChevronRight data-icon="inline-end" />
      </Button>
      <Button variant="outline" size="icon" aria-label="Add"><Plus /></Button>
    </div>
  ),
};

/**
 * Primary-button contrast explorations. The default button now uses
 * `--foreground` on `--background` (black on white in light, inverse in dark),
 * which clears AAA in both directions. The alternatives below explore using
 * the brand sun color as the primary button fill instead — each is the real
 * Button with its colors overridden via arbitrary-value classes (twMerge lets
 * later bg/text utilities win). Ratios are measured against the light theme;
 * switch the Theme toolbar to see how each carries into dark mode.
 */
export const ContrastProposals: Story = {
  render: () => (
    <div className="flex flex-col gap-8">
      <Proposal
        label="Default (foreground/background)"
        description="Black bg + white text in light, inverse in dark. Neutral, high-contrast, no brand color."
        ratio="19.68:1 (AAA)"
        className="bg-foreground text-background hover:bg-foreground/90"
      />
      <Proposal
        label="Original primary"
        description="Bright sun + near-black. ~8.5:1 (AAA), but reads heavy/industrial."
        ratio="8.54:1"
        className="bg-primary text-primary-foreground hover:bg-primary/80"
      />
      <Proposal
        label="A · Solid sun + white"
        description="Same sun, one stop darker, white type. Clean confident CTA, preserves brand hue."
        ratio="4.66:1 (AA)"
        className="bg-[oklch(0.58_0.19_39)] text-[oklch(0.99_0.005_79)] hover:bg-[oklch(0.52_0.19_39)]"
      />
      <Proposal
        label="B · Bright sun + tinted-dark fg"
        description="Keep the bright fill, tint the text toward the brand hue instead of pure black."
        ratio="7.14:1 (AAA)"
        className="bg-primary text-[oklch(0.25_0.10_39)] hover:bg-primary/80"
      />
      <Proposal
        label="C · Outline + darker sun text"
        description="Border-only, primary text in a darker sun so it clears AA on white."
        ratio="5.50:1 (AA)"
        className="border-[oklch(0.54_0.19_39)] bg-transparent text-[oklch(0.54_0.19_39)] hover:bg-[oklch(0.54_0.19_39/0.08)]"
      />
      <Proposal
        label="D · Softened fill + black"
        description="Lower chroma fill (calm tint) with the existing near-black text. Quieter."
        ratio="8.23:1 (AAA)"
        className="bg-[oklch(0.74_0.10_39)] text-primary-foreground hover:bg-[oklch(0.70_0.10_39)]"
      />
    </div>
  ),
};

function Proposal({
  label,
  description,
  ratio,
  className,
}: {
  label: string;
  description: string;
  ratio: string;
  className: string;
}) {
  return (
    <div className="flex flex-col gap-3 rounded-lg border border-border bg-background p-4">
      <div className="flex items-baseline justify-between gap-4">
        <h3 className="text-sm font-semibold text-foreground">{label}</h3>
        <span className="rounded-full bg-muted px-2 py-0.5 font-mono text-xs text-muted-foreground">
          {ratio}
        </span>
      </div>
      <p className="text-sm text-muted-foreground">{description}</p>
      <div className="flex flex-wrap items-center gap-3 pt-1">
        <Button className={className}>Subscribe</Button>
        <Button className={className} size="lg">Get started</Button>
        <Button className={className} size="sm">Save</Button>
        <Button className={className} size="icon" aria-label="Add"><Plus /></Button>
      </div>
    </div>
  );
}
