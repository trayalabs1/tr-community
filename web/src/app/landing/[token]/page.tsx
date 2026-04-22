import { notFound } from "next/navigation";
import { LandingScreen } from "@/screens/auth/LandingScreen/LandingScreen";

type Props = {
  params: Promise<{
    token: string;
  }>;
  searchParams: Promise<{
    share?: string;
    streak_count?: string;
    reward_coins?: string;
  }>;
};

export default async function Page(props: Props) {
  const { token } = await props.params;
  const searchParams = await props.searchParams;

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

  return (
    <LandingScreen
      token={cleanToken}
      share={searchParams?.share === "true"}
      streakCount={searchParams?.streak_count ? parseInt(searchParams.streak_count, 10) : undefined}
      rewardCoins={searchParams?.reward_coins ? parseInt(searchParams.reward_coins, 10) : undefined}
    />
  );
}
