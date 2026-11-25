import { Account } from "@/api/openapi-schema";

type Props = {
  account?: Account;
};

export function MiniProfileAvatar({ account }: Props) {
  const initial = (
    account?.name?.charAt(0) ||
    account?.handle?.charAt(0) ||
    "U"
  ).toUpperCase();

  return (
    <span
      style={{
        position: "relative",
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        flexShrink: 0,
        width: 18,
        height: 18,
        overflow: "hidden",
        borderRadius: "50%",
        backgroundColor: "#ffdd81",
        color: "#404040",
        fontSize: 8.5,
        lineHeight: "10px",
        fontWeight: 600,
      }}
    >
      {initial}
      <span
        style={{
          position: "absolute",
          bottom: 0,
          left: "50%",
          transform: "translateX(-50%)",
          padding: "0.5px 4.7px",
          backgroundColor: "rgba(0,0,0,0.6)",
          color: "#ffffff",
          fontSize: 3.8,
          lineHeight: "4.7px",
          letterSpacing: "0.38px",
          fontWeight: 600,
        }}
      >
        YOU
      </span>
    </span>
  );
}
