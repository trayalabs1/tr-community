import { Button } from "@/components/ui/button";
import { FormControl } from "@/components/ui/form/FormControl";
import { FormHelperText } from "@/components/ui/form/FormHelperText";
import { FormLabel } from "@/components/ui/form/FormLabel";
import { NumberInputField } from "@/components/ui/form/NumberInputField";
import { Heading } from "@/components/ui/heading";
import { CardBox, Flex, WStack, styled } from "@/styled-system/jsx";
import { lstack } from "@/styled-system/patterns";

import { Props, useFeedRankingSettings } from "./useFeedRankingSettings";

const decimalFormat: Intl.NumberFormatOptions = { maximumFractionDigits: 3 };

export function FeedRankingSettingsForm(props: Props) {
  const { control, formState, onSubmit } = useFeedRankingSettings(props);

  return (
    <styled.form
      width="full"
      display="flex"
      flexDirection="column"
      gap="4"
      onSubmit={onSubmit}
    >
      <CardBox className={lstack()}>
        <WStack>
          <Heading size="md">Feed ranking settings</Heading>
          <Button type="submit" loading={formState.isSubmitting}>
            Save
          </Button>
        </WStack>

        <Heading size="sm">Content score weights</Heading>
        <FormHelperText>
          Applied once when a post is scored by the LLM. Changing these only
          affects newly-scored posts, not the rank_score of existing posts.
        </FormHelperText>

        <Flex flexDir={{ base: "column", md: "row" }} gap="2">
          <FormControl>
            <FormLabel>Positivity weight</FormLabel>
            <NumberInputField
              control={control}
              name="wPositivity"
              scrubber={true}
              min={0}
              max={10}
              step={0.1}
              formatOptions={decimalFormat}
            />
          </FormControl>

          <FormControl>
            <FormLabel>Feed value weight</FormLabel>
            <NumberInputField
              control={control}
              name="wFeedValue"
              scrubber={true}
              min={0}
              max={10}
              step={0.1}
              formatOptions={decimalFormat}
            />
          </FormControl>

          <FormControl>
            <FormLabel>Quality weight</FormLabel>
            <NumberInputField
              control={control}
              name="wQuality"
              scrubber={true}
              min={0}
              max={10}
              step={0.1}
              formatOptions={decimalFormat}
            />
          </FormControl>

          <FormControl>
            <FormLabel>Category boost weight</FormLabel>
            <NumberInputField
              control={control}
              name="wCategory"
              scrubber={true}
              min={0}
              max={10}
              step={0.1}
              formatOptions={decimalFormat}
            />
          </FormControl>
        </Flex>

        <Heading size="sm">Live read-time multipliers</Heading>
        <FormHelperText>
          Applied every time the feed is read. Changes take effect on the
          very next request.
        </FormHelperText>

        <Flex flexDir={{ base: "column", md: "row" }} gap="2">
          <FormControl>
            <FormLabel>Freshness halflife (hours)</FormLabel>
            <NumberInputField
              control={control}
              name="freshnessHalflifeHours"
              scrubber={true}
              min={0.1}
              max={24 * 30}
              step={1}
            />
            <FormHelperText>
              How many hours it takes for a post&apos;s freshness score to
              decay.
            </FormHelperText>
          </FormControl>

          <FormControl>
            <FormLabel>Image post multiplier</FormLabel>
            <NumberInputField
              control={control}
              name="formatMultiplier"
              scrubber={true}
              min={0}
              max={10}
              step={0.1}
              formatOptions={decimalFormat}
            />
            <FormHelperText>
              Multiplier applied to posts with at least one image attached.
            </FormHelperText>
          </FormControl>

          <FormControl>
            <FormLabel>Negative sentiment multiplier</FormLabel>
            <NumberInputField
              control={control}
              name="sentimentMultiplier"
              scrubber={true}
              min={0}
              max={10}
              step={0.05}
              formatOptions={decimalFormat}
            />
            <FormHelperText>
              Multiplier applied to posts with negative sentiment.
            </FormHelperText>
          </FormControl>
        </Flex>

        <Heading size="sm">Daily engagement job weights</Heading>
        <FormHelperText>
          Applied once a day by a scheduled job that folds likes/replies from
          the last 24 hours into each post&apos;s rank_score.
        </FormHelperText>

        <Flex flexDir={{ base: "column", md: "row" }} gap="2">
          <FormControl>
            <FormLabel>Like weight</FormLabel>
            <NumberInputField
              control={control}
              name="likeWeight"
              scrubber={true}
              min={0}
              max={100}
              step={0.5}
              formatOptions={decimalFormat}
            />
          </FormControl>

          <FormControl>
            <FormLabel>Reply weight</FormLabel>
            <NumberInputField
              control={control}
              name="replyWeight"
              scrubber={true}
              min={0}
              max={100}
              step={0.5}
              formatOptions={decimalFormat}
            />
          </FormControl>
        </Flex>
      </CardBox>
    </styled.form>
  );
}
