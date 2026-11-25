import { test } from "uvu";
import * as assert from "uvu/assert";

import {
  buildGenericBody,
  escapeHtml,
  isSafeMediaUrl,
  parseMediaParam,
} from "./sharePostContent";

test("body with script tag is escaped, not markup", () => {
  const out = buildGenericBody("<script>alert(1)</script>hi", []);
  assert.is(out, "<p>&lt;script&gt;alert(1)&lt;/script&gt;hi</p>");
  assert.not.match(out, /<script>/);
});

test("escapeHtml covers the five entities", () => {
  assert.is(escapeHtml(`&<>"'`), "&amp;&lt;&gt;&quot;&#39;");
});

test("javascript: and data: media urls are rejected", () => {
  assert.is(isSafeMediaUrl("javascript:alert(1)"), false);
  assert.is(isSafeMediaUrl("data:text/html,<script>"), false);
  assert.is(isSafeMediaUrl("not a url"), false);
  assert.is(isSafeMediaUrl("https://cdn.traya.health/a.jpg"), true);
  assert.is(isSafeMediaUrl("http://cdn.traya.health/a.jpg"), true);
});

test("media splits on comma, drops empties and whitespace", () => {
  assert.equal(parseMediaParam("a,,b, c ,"), ["a", "b", "c"]);
  assert.equal(parseMediaParam(""), []);
  assert.equal(parseMediaParam(undefined), []);
});

test("three media urls become three img tags", () => {
  const media = ["https://x/a.jpg", "https://x/b.jpg", "https://x/c.jpg"];
  assert.is(
    buildGenericBody("my text", media),
    '<p>my text</p><img src="https://x/a.jpg" alt="" /><img src="https://x/b.jpg" alt="" /><img src="https://x/c.jpg" alt="" />',
  );
});

test("quote in media url cannot break out of the src attribute", () => {
  const out = buildGenericBody("t", ['https://x/a.jpg" onerror="alert(1)']);
  assert.not.match(out, /onerror="alert/);
  assert.match(out, /&quot;/);
});

test("share query round-trips a body containing & and #", () => {
  const params = new URLSearchParams();
  params.set("body", "tea & biscuits #win");
  params.set("media", "https://x/a.jpg,https://x/b.jpg");
  const parsed = new URLSearchParams(params.toString());
  assert.is(parsed.get("body"), "tea & biscuits #win");
  assert.equal(parseMediaParam(parsed.get("media") ?? undefined), [
    "https://x/a.jpg",
    "https://x/b.jpg",
  ]);
});

test.run();
