#!/bin/bash

COOKIE="storyden-session=d7bpr2b8tclc73c932hg"
BASE_URL="https://cbe.traya.health/api/channels"
BATCH_SIZE=50
DELAY_BETWEEN_BATCHES=60
DELAY_BETWEEN_CHANNELS=60

CHANNELS=(
  "d5suajie1bac73fhst0g:Month 5 Heroines:114"
  "d5su9uqe1bac73fhsstg:Month 5 Heroines:114"
  "d5sua4ae1bac73fhssug:Month 6-8 Heroines:156"
  "d5su9pqe1bac73fhsssg:Month 4 Heroines:185"
  "d5su9hqe1bac73fhssrg:Month 3 Heroines:230"
  "d5su8tie1bac73fhssog:Month 5 Warriors:254"
  "d5su8iae1bac73fhssmg:Month 3 Warriors:306"
  "d5su562e1bac73fhssk0:Month 2 Heroines:360"
  "d5su8o2e1bac73fhssng:Month 4 Warriors:372"
  "d5su94qe1bac73fhsspg:Month 6-8 Warriors:489"
  "d5su39ae1bac73fhssdg:Month 1 Heroines:573"
  "d6alr8lfross7398soqg:Month 2 Warriors:706"
  "d5su2pqe1bac73fhsscg:Month 1 Warriors:1247"
)

echo "=== Scoring & Ranking Script ==="
echo "Batch size: $BATCH_SIZE | Delay between batches: ${DELAY_BETWEEN_BATCHES}s | Delay between channels: ${DELAY_BETWEEN_CHANNELS}s"
echo ""

TOTAL_SCORED=0

for entry in "${CHANNELS[@]}"; do
  IFS=':' read -r CHANNEL_ID NAME POST_COUNT <<< "$entry"

  echo "--- Channel: $NAME ($CHANNEL_ID) | ~$POST_COUNT posts ---"

  CHANNEL_TOTAL=0
  BATCH_NUM=0

  while true; do
    BATCH_NUM=$((BATCH_NUM + 1))

    RESPONSE=$(curl -s -X POST "${BASE_URL}/${CHANNEL_ID}/scoring/score-unscored?limit=${BATCH_SIZE}&include_failed=true" \
      --header "Cookie: ${COOKIE}" \
      --header "Accept: application/json")

    ENQUEUED=$(echo "$RESPONSE" | grep -o '"posts_enqueued":[0-9]*' | grep -o '[0-9]*')
    SUCCESS=$(echo "$RESPONSE" | grep -o '"success":true')

    if [ -z "$SUCCESS" ]; then
      echo "  ERROR: Unexpected response: $RESPONSE"
      echo "  Skipping channel."
      break
    fi

    CHANNEL_TOTAL=$((CHANNEL_TOTAL + ENQUEUED))
    echo "  Batch $BATCH_NUM: enqueued $ENQUEUED posts (channel total: $CHANNEL_TOTAL)"

    if [ "$ENQUEUED" -eq 0 ]; then
      echo "  Done. Total enqueued for $NAME: $CHANNEL_TOTAL"
      break
    fi

    echo "  Waiting ${DELAY_BETWEEN_BATCHES}s before next batch..."
    sleep $DELAY_BETWEEN_BATCHES
  done

  TOTAL_SCORED=$((TOTAL_SCORED + CHANNEL_TOTAL))

  echo ""
  echo "  Waiting ${DELAY_BETWEEN_CHANNELS}s before next channel..."
  sleep $DELAY_BETWEEN_CHANNELS
  echo ""
done

echo "=== Complete ==="
echo "Total posts enqueued across all channels: $TOTAL_SCORED"
