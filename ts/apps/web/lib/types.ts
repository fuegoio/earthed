import type { ApiToken, TokenOutputBody } from "@planetary/api-client"

export type {
  ApiToken,
  Category,
  CategoryWritable,
  Enclosure,
  Entry,
  EntryWritable,
  ErrorDetail,
  ErrorModel,
  Feed,
  FeedWritable,
  PreviewFeedBody,
  PreviewFeedItem,
  User,
} from "@planetary/api-client"

// Web-friendly aliases for generated types whose generated names are awkward
// to use in components.
export type APIToken = ApiToken
export type CreatedToken = TokenOutputBody
