import { SharePostScreen } from "@/screens/compose/SharePostScreen";

type Props = {
  params: Promise<{
    id: string;
  }>;
  searchParams: Promise<{
    streak_count?: string;
    reward_coins?: string;
    category?: string;
    type?: string;
    score?: string;
    text?: string;
    image?: string;
  }>;
};

const MAX_SHARE_TEXT_LENGTH = 500;

function parseShareText(raw?: string): string | undefined {
  const text = raw?.trim();
  if (!text) return undefined;
  return text.slice(0, MAX_SHARE_TEXT_LENGTH);
}

function parseShareImage(raw?: string): string | undefined {
  if (!raw) return undefined;
  try {
    const url = new URL(raw);
    return url.protocol === "https:" ? url.toString() : undefined;
  } catch {
    return undefined;
  }
}

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
  const parsedScore = searchParams?.score
    ? parseInt(searchParams.score, 10)
    : NaN;
  const score =
    Number.isFinite(parsedScore) && parsedScore >= 0 && parsedScore <= 100
      ? parsedScore
      : undefined;

  return (
    <SharePostScreen
      channelID={id}
      streakCount={streakCount}
      rewardCoins={rewardCoins}
      category={searchParams?.category}
      type={searchParams?.type}
      score={score}
      text={parseShareText(searchParams?.text)}
      image={parseShareImage(searchParams?.image)}
    />
  );
}
