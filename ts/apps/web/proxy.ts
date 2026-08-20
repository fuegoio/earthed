import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const PUBLIC_ROUTES = ["/login", "/signup", "/~offline"];

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const session = request.cookies.get("limen_session");

  const isPublicRoute = PUBLIC_ROUTES.includes(pathname);

  // Gate private routes behind the session cookie. We do NOT bounce
  // logged-in users off public routes here: the session cookie only proves the
  // browser has a cookie, not that it's valid (it may be expired, or the API
  // may be down). Forcing /login → / here would loop with the layout's
  // redirect("/login") when the API rejects the session.
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
