"use client";

import Link from "next/link";

import { Account } from "@/api/openapi-schema";
import { styled } from "@/styled-system/jsx";

type Props = {
  account: Account;
  /** Wrap the avatar in a link to the member's profile. */
  asLink?: boolean;
};

export function SelfAvatarBadge({ account, asLink = true }: Props) {
  const initial = (
    account.name?.charAt(0) || account.handle.charAt(0)
  ).toUpperCase();

  const avatar = (
    <styled.div
      position="relative"
      overflow="hidden"
      display="flex"
      alignItems="center"
      justifyContent="center"
      flexShrink="0"
      rounded="full"
      style={{ width: "38px", height: "38px", backgroundColor: "#ffdd81" }}
    >
      <styled.span
        fontWeight="semibold"
        style={{ color: "#404040", fontSize: "18px", lineHeight: "22px" }}
      >
        {initial}
      </styled.span>
      <styled.div
        position="absolute"
        display="flex"
        alignItems="center"
        justifyContent="center"
        style={{
          bottom: "0",
          left: "50%",
          transform: "translateX(-50%)",
          padding: "1px 10px",
          backgroundColor: "rgba(0,0,0,0.6)",
        }}
      >
        <styled.span
          fontWeight="semibold"
          style={{
            color: "white",
            fontSize: "8px",
            lineHeight: "10px",
            letterSpacing: "0.8px",
          }}
        >
          YOU
        </styled.span>
      </styled.div>
    </styled.div>
  );

  if (!asLink) {
    return avatar;
  }

  return (
    <Link href={`/m/${account.handle}`} style={{ textDecoration: "none" }}>
      {avatar}
    </Link>
  );
}
