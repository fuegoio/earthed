import type { ApiToken, TokenOutputBody } from "@sunred/api-client";

export type {
  ApiToken,
  Enclosure,
  Entry,
  EntryWritable,
  ErrorDetail,
  ErrorModel,
  Feed,
  FeedList,
  FeedListFeed,
  FeedListWritable,
  FeedSubscribersResponse,
  FeedWritable,
  Folder,
  ImportFeedListResponse,
  PreviewFeedBody,
  PreviewFeedItem,
  PublicProfileResponse,
  SharedArticle,
  User,
  UserProfile,
} from "@sunred/api-client";

// Web-friendly aliases for generated types whose generated names are awkward
// to use in components.
export type APIToken = ApiToken;
export type CreatedToken = TokenOutputBody;
