import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { handle } from "@/api/client";
import { useSettingsMutation } from "@/lib/settings/mutation";
import { AdminSettings } from "@/lib/settings/settings";

export type Props = {
  settings: AdminSettings;
};

export const FormSchema = z.object({
  wPositivity: z.number().min(0).max(10),
  wFeedValue: z.number().min(0).max(10),
  wQuality: z.number().min(0).max(10),
  wCategory: z.number().min(0).max(10),
  freshnessHalflifeHours: z.number().min(0.1).max(24 * 30),
  formatMultiplier: z.number().min(0).max(10),
  sentimentMultiplier: z.number().min(0).max(10),
  likeWeight: z.number().min(0).max(100),
  replyWeight: z.number().min(0).max(100),
});
export type Form = z.infer<typeof FormSchema>;

export function useFeedRankingSettings({ settings }: Props) {
  const { revalidate, updateSettings } = useSettingsMutation();
  const feedRanking = settings.services?.feed_ranking;

  const form = useForm<Form>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      wPositivity: feedRanking?.w_positivity ?? 1.0,
      wFeedValue: feedRanking?.w_feed_value ?? 1.5,
      wQuality: feedRanking?.w_quality ?? 1.0,
      wCategory: feedRanking?.w_category ?? 1.0,
      freshnessHalflifeHours: feedRanking?.freshness_halflife_hours ?? 24,
      formatMultiplier: feedRanking?.format_multiplier ?? 2.5,
      sentimentMultiplier: feedRanking?.sentiment_multiplier ?? 0.1,
      likeWeight: feedRanking?.like_weight ?? 2.0,
      replyWeight: feedRanking?.reply_weight ?? 0.5,
    },
  });

  const onSubmit = form.handleSubmit(async (data) => {
    await handle(
      async () => {
        await updateSettings({
          services: {
            feed_ranking: {
              w_positivity: data.wPositivity,
              w_feed_value: data.wFeedValue,
              w_quality: data.wQuality,
              w_category: data.wCategory,
              freshness_halflife_hours: data.freshnessHalflifeHours,
              format_multiplier: data.formatMultiplier,
              sentiment_multiplier: data.sentimentMultiplier,
              like_weight: data.likeWeight,
              reply_weight: data.replyWeight,
            },
          },
        });
      },
      {
        promiseToast: {
          loading: "Saving settings...",
          success: "Settings saved",
        },
        cleanup: async () => {
          await revalidate();
        },
      },
    );
  });

  return {
    register: form.register,
    control: form.control,
    formState: form.formState,
    onSubmit,
  };
}
