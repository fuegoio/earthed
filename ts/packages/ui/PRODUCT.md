# Product

## Register

product

## Platform

web

## Users

Developers building Sunred's web app and docs site. They consume the
design system's components, tokens, and prose styles to assemble product
surfaces. They value consistency, type safety, and components that work
without configuration.

## Product Purpose

The shared UI package (`@workspace/ui`) provides the visual vocabulary for
Sunred's frontend surfaces: color tokens, typography, spacing, component
primitives (shadcn/ui on `@base-ui/react`), and the typeset prose system.
Success looks like every Sunred surface sharing the same component
vocabulary, the same color logic, and the same reading-first defaults.

## Positioning

A modern, fast, developer-friendly RSS reader with a clean API and multiple
interfaces. The complete package, not a single differentiator.

## Brand Personality

Warm and approachable. The design system should produce interfaces that feel
like a comfortable reading space, not a data dashboard. Clean, calm,
generous. The system itself should feel crafted — thoughtful defaults,
consistent affordances, no unnecessary surface area.

## Anti-references

Design systems that produce generic SaaS dashboards: cream backgrounds,
identical card grids, flat blue palettes, components that look like they
came from a template rather than a product with a point of view.

## Design Principles

- **Reading first.** Every component and token should serve the reading
  experience. Buttons, forms, and layout should recede behind content.
- **Warm, not clinical.** Color, spacing, and typography should produce a
  comfortable surface, not a sterile one. The warm/approachable personality
  starts here, in the tokens.
- **Complete, not cluttered.** Ship the components the product needs, each
  with full state coverage (hover, focus, active, disabled, loading, error).
  Don't ship components nobody uses.
- **Modern craft.** OKLCH color, Tailwind v4, shadcn/ui on base-ui. The
  system should feel current and maintained, not inherited debt.
- **Accessibility is baseline.** WCAG 2.2 AA. Components must meet contrast,
  keyboard, and screen reader requirements by default. Reduced motion is
  respected at the token level.

## Accessibility & Inclusion

WCAG 2.2 AA. All components must be keyboard-accessible, screen-reader-
compatible, and meet AA contrast ratios. Focus states are visible and
consistent. Reduced-motion preferences are respected at the system level.
