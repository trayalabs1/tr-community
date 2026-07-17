import { test } from "uvu";
import * as assert from "uvu/assert";

import { containsMobileNumber } from "./content-validation";

test("plain text with no digits", () => {
  assert.is(containsMobileNumber("hello world"), false);
});

test("five or fewer digits allowed", () => {
  assert.is(containsMobileNumber("order 12345"), false);
});

test("digits split across html allowed when five or fewer", () => {
  assert.is(containsMobileNumber("<p>call at 3pm on day 12</p>"), false);
});

test("ten digit number blocked", () => {
  assert.is(containsMobileNumber("9876543210"), true);
});

test("number with spaces blocked", () => {
  assert.is(containsMobileNumber("call me on 98765 43210"), true);
});

test("number inside html body blocked", () => {
  assert.is(containsMobileNumber("<p>my number is 9876543210</p>"), true);
});

test("digits inside html tags and attributes ignored", () => {
  assert.is(containsMobileNumber('<a href="/t/123456789">link</a>'), false);
});

test("empty string allowed", () => {
  assert.is(containsMobileNumber(""), false);
});

test.run();
