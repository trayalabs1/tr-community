import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAccountGet } from "src/api/openapi-client/accounts";
import { useMoengage } from "@/lib/moengage/useMoengage";

const PRIVATE_PAGES = ["/settings", "/new", "/admin"];

function privatePage(pathName: string): boolean {
  return PRIVATE_PAGES.includes(pathName);
}

export function useAuthProvider() {
  const { isLoading, data, error } = useAccountGet();
  const { push } = useRouter();
  const pathname = usePathname();
  const { setUserAttributes } = useMoengage();

  const loggedIn = Boolean(data) && !error;
  const isPrivate = pathname && privatePage(pathname);

  useEffect(() => {
    if (loggedIn && data) {
      setUserAttributes({
        name: data.name,
        email: data.email_addresses?.[0]?.email_address,
      });
    }
  }, [loggedIn, data, setUserAttributes]);

  useEffect(() => {
    if (isLoading) return;

    if (!loggedIn && isPrivate) {
      push("/register");
    }
    if (loggedIn && (pathname === "/register" || pathname === "/login")) {
      push("/");
    }
  }, [isLoading, loggedIn, isPrivate, pathname, push]);

  return;
}
