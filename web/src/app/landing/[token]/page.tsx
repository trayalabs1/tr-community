import { notFound } from "next/navigation";
import { LandingScreen } from "@/screens/auth/LandingScreen/LandingScreen";

type Props = {
  params: Promise<{
    token: string;
  }>;
};

export default async function Page(props: Props) {
  const { token } = await props.params;

  if (!token || typeof token !== "string") {
    notFound();
  }

  const decodedToken = decodeURIComponent(token);
  const cleanToken = decodedToken.includes("&")
    ? decodedToken.split("&")[0]?.trim() ?? ""
    : decodedToken.split("?")[0]?.trim() ?? "";

  if (!cleanToken) {
    notFound();
  }

  return <LandingScreen token={cleanToken} />;
}
