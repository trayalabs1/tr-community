import { test } from "uvu";
import * as assert from "uvu/assert";

import { hasChannelAccess } from "./locateAccess";

const CHANNELS = [
  { id: "chan_alpha", name: "Month 1 Heroines", slug: "month-1-heroines" },
  { id: "chan_beta", name: "Traya Explorers", slug: "traya-explorers" },
];

test("member of the channel has access", () => {
  assert.is(hasChannelAccess("chan_alpha", CHANNELS), true);
});

test("non-member of the channel has no access", () => {
  assert.is(hasChannelAccess("chan_forbidden", CHANNELS), false);
});

test("no accessible channels means no access", () => {
  assert.is(hasChannelAccess("chan_alpha", []), false);
});

test("missing channel id means no access", () => {
  assert.is(hasChannelAccess(undefined, CHANNELS), false);
});

test("empty channel id means no access", () => {
  assert.is(hasChannelAccess("", CHANNELS), false);
});

// A thread shared into a channel the viewer can reach is reachable, even when
// its home channel is not.
test("thread shared into an accessible channel has access", () => {
  assert.is(
    hasChannelAccess("chan_forbidden", CHANNELS, ["chan_alpha"]),
    true,
  );
});

test("shared into an inaccessible channel still has no access", () => {
  assert.is(
    hasChannelAccess("chan_forbidden", CHANNELS, ["chan_also_forbidden"]),
    false,
  );
});

test("home channel access wins even with no shares", () => {
  assert.is(hasChannelAccess("chan_alpha", CHANNELS, []), true);
});

test("shared channel list is optional", () => {
  assert.is(hasChannelAccess("chan_forbidden", CHANNELS, undefined), false);
});

test("missing home channel but accessible share still grants access", () => {
  assert.is(hasChannelAccess(undefined, CHANNELS, ["chan_beta"]), true);
});

test.run();
