import { SharePostScreen } from "@/screens/compose/SharePostScreen";
import { parseMediaParam } from "@/screens/compose/sharePostContent";

type Props = {
  params: Promise<{
    id: string;
  }>;
  searchParams: Promise<{
    streak_count?: string;
    reward_coins?: string;
    category?: string;
    type?: string;
    body?: string;
    media?: string;
  }>;
};

export default async function Page(props: Props) {
  const { id } = await props.params;
  const searchParams = await props.searchParams;

  const parsedStreak = searchParams?.streak_count
    ? parseInt(searchParams.streak_count, 10)
    : NaN;
  const parsedCoins = searchParams?.reward_coins
    ? parseInt(searchParams.reward_coins, 10)
    : NaN;
  const streakCount = Number.isFinite(parsedStreak) ? parsedStreak : undefined;
  const rewardCoins = Number.isFinite(parsedCoins) ? parsedCoins : undefined;

  return (
    <SharePostScreen
      channelID={id}
      streakCount={streakCount}
      rewardCoins={rewardCoins}
      category={searchParams?.category}
      type={searchParams?.type}
      body={searchParams?.body}
      media={parseMediaParam(searchParams?.media)}
    />
  );
}
