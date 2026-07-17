import { test } from "uvu";
import * as assert from "uvu/assert";

import { containsMobileNumber } from "./content-validation";

test("plain text with no digits", () => {
  assert.is(containsMobileNumber("hello world"), false);
});

test("short number allowed", () => {
  assert.is(containsMobileNumber("order 12345"), false);
});

test("nine digits allowed", () => {
  assert.is(containsMobileNumber("987654321"), false);
});

test("bare ten digit number blocked", () => {
  assert.is(containsMobileNumber("9876543210"), true);
});

test("ten digits with country code blocked", () => {
  assert.is(containsMobileNumber("+919876543210"), true);
});

test("ten digits with single-digit country code blocked", () => {
  assert.is(containsMobileNumber("+19876543210"), true);
});

test("ten digit number embedded in text blocked", () => {
  assert.is(containsMobileNumber("call me 9876543210 now"), true);
});

test("ten digit number inside html body blocked", () => {
  assert.is(containsMobileNumber("<p>my number is 9876543210</p>"), true);
});

test("number with spaces allowed", () => {
  assert.is(containsMobileNumber("call me on 98765 43210"), false);
});

test("eleven digit run allowed", () => {
  assert.is(containsMobileNumber("98765432101"), false);
});

test("twelve digit id allowed", () => {
  assert.is(containsMobileNumber("id 987654321012"), false);
});

test("digits inside html tags and attributes ignored", () => {
  assert.is(containsMobileNumber('<a href="/t/1234567890">link</a>'), false);
});

test("empty string allowed", () => {
  assert.is(containsMobileNumber(""), false);
});

test.run();
