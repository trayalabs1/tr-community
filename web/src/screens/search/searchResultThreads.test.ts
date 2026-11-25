import { test } from "uvu";
import * as assert from "uvu/assert";

import { DatagraphItemKind } from "@/api/openapi-schema";

import { searchResultThreads } from "./searchResultThreads";

const threadItem = {
  kind: DatagraphItemKind.thread,
  ref: { id: "t1", slug: "thread-one", reply_status: { replied: 0, replies: 2 } },
};

const postItem = {
  kind: DatagraphItemKind.post,
  ref: { id: "p1", slug: "post-one" },
};

const profileItem = {
  kind: DatagraphItemKind.profile,
  ref: { id: "u1", handle: "someone" },
};

test("returns an empty array when items are undefined", () => {
  assert.equal(searchResultThreads(undefined), []);
});

test("keeps thread-kind items", () => {
  const result = searchResultThreads([threadItem] as never);
  assert.is(result.length, 1);
  assert.is(result[0]?.id, "t1");
});

test("drops post-kind items because Post lacks reply_status", () => {
  assert.equal(searchResultThreads([postItem] as never), []);
});

test("drops non-thread kinds and preserves order of the rest", () => {
  const result = searchResultThreads([postItem, threadItem, profileItem] as never);
  assert.equal(
    result.map((t) => t.id),
    ["t1"],
  );
});

test.run();
