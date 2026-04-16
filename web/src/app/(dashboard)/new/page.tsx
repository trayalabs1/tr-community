"use client";

import { use } from "react";
import { z } from "zod";

import { ComposeScreen } from "src/screens/compose/ComposeScreen";

const QuerySchema = z.object({
  id: z.string().optional(),
  channel: z.string().optional(),
  category: z.string().optional(),
  streak_count: z.string().optional(),
  reward_coins: z.string().optional(),
});

type Props = {
  searchParams: Promise<z.infer<typeof QuerySchema>>;
};

export default function Page(props: Props) {
  const searchParams = use(props.searchParams);
  const params = QuerySchema.parse(searchParams);
  return (
    <ComposeScreen
      editing={params.id}
      channelID={params.channel}
      categoryID={params.category}
      streakCount={params.streak_count ? parseInt(params.streak_count, 10) : undefined}
      rewardCoins={params.reward_coins ? parseInt(params.reward_coins, 10) : undefined}
    />
  );
}
