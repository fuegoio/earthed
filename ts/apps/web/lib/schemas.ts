import { z } from "zod";

export const handleSchema = z.object({
  handle: z
    .string()
    .min(1, "Enter your AT Proto handle")
    .max(253, "Handle is too long")
    .refine((v) => !v.includes(" "), "Handle cannot contain spaces"),
});

export type HandleValues = z.infer<typeof handleSchema>;

export const subscribeFeedSchema = z.object({
  feed_url: z
    .string()
    .min(1, "Enter a URL")
    .max(2048, "URL is too long")
    .refine((val) => {
      try {
        // Accept URLs with or without a scheme.
        const withScheme = /^https?:\/\//i.test(val) ? val : `https://${val}`;
        new URL(withScheme);
        return true;
      } catch {
        return false;
      }
    }, "Enter a valid URL"),
});

export function normalizeFeedURL(input: string): string {
  const trimmed = input.trim();
  if (/^https?:\/\//i.test(trimmed)) return trimmed;
  return `https://${trimmed}`;
}

export type SubscribeFeedValues = z.infer<typeof subscribeFeedSchema>;
