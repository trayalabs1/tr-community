const MAX_DIGITS = 5;

export const MOBILE_NUMBER_ERROR = "Don't share mobile numbers";

export function containsMobileNumber(body: string): boolean {
  const text = body.replace(/<[^>]*>/g, "");
  const digitCount = (text.match(/\d/g) ?? []).length;
  return digitCount > MAX_DIGITS;
}
