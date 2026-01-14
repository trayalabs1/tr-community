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

  const cleanToken = token.split("&")[0]?.split("?")[0]?.trim() ?? "";

  if (!cleanToken) {
    notFound();
  }

  return <LandingScreen token={cleanToken} />;
}
