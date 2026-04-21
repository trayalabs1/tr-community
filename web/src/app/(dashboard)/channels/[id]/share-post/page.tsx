import { SharePostScreen } from "@/screens/compose/SharePostScreen";

type Props = {
  params: Promise<{
    id: string;
  }>;
  searchParams: Promise<{
    streak_count?: string;
    reward_coins?: string;
  }>;
};

export default async function Page(props: Props) {
  const { id } = await props.params;
  const searchParams = await props.searchParams;

  const streakCount = searchParams?.streak_count
    ? parseInt(searchParams.streak_count, 10)
    : undefined;
  const rewardCoins = searchParams?.reward_coins
    ? parseInt(searchParams.reward_coins, 10)
    : undefined;

  return (
    <SharePostScreen
      channelID={id}
      streakCount={streakCount}
      rewardCoins={rewardCoins}
    />
  );
}
