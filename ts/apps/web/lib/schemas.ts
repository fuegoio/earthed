import { z } from "zod"

export const signinSchema = z.object({
  email: z.string().email("Enter a valid email address"),
  password: z.string().min(1, "Password is required"),
})

export const signupSchema = z
  .object({
    email: z.string().email("Enter a valid email address"),
    password: z
      .string()
      .min(8, "Password must be at least 8 characters")
      .regex(/[A-Z]/, "Password must contain an uppercase letter")
      .regex(/[a-z]/, "Password must contain a lowercase letter")
      .regex(/[0-9]/, "Password must contain a number"),
    confirmPassword: z.string().min(1, "Please confirm your password"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  })

export type SigninValues = z.infer<typeof signinSchema>
export type SignupValues = z.infer<typeof signupSchema>

export const subscribeFeedSchema = z.object({
  feed_url: z
    .string()
    .min(1, "Enter a feed URL")
    .url("Enter a valid URL")
    .max(2048, "URL is too long"),
})

export type SubscribeFeedValues = z.infer<typeof subscribeFeedSchema>
