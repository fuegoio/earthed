import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const PUBLIC_ROUTES = ["/login", "/signup"];

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const session = request.cookies.get("limen_session");

  const isPublicRoute = PUBLIC_ROUTES.includes(pathname);

  if (isPublicRoute && session) {
    return NextResponse.redirect(new URL("/", request.url));
  }

  if (!isPublicRoute && !session) {
    const loginUrl = new URL("/login", request.url);
    // Preserve the full path + query so the user returns to the exact page
    // they requested (e.g. /device?user_code=PLN-XXXX-XXXX).
    loginUrl.searchParams.set(
      "redirect",
      request.nextUrl.pathname + request.nextUrl.search,
    );
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next|favicon.ico|.*\\.).*)"],
};
