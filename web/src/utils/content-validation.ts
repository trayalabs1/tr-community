export const MOBILE_NUMBER_ERROR = "Don't share mobile numbers";

// Matches a phone number: an optional country code (+ and 1-3 digits) followed
// by exactly 10 digits with no separators, not adjacent to other digits.
const MOBILE_NUMBER_PATTERN = /(?<!\d)(?:\+\d{1,3})?\d{10}(?!\d)/;

export function containsMobileNumber(body: string): boolean {
  const text = body.replace(/<[^>]*>/g, "");
  return MOBILE_NUMBER_PATTERN.test(text);
}
