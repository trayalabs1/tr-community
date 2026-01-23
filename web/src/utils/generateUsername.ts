import { Config, uniqueNamesGenerator, adjectives, names, colors } from "unique-names-generator";

export function generateRandomUsername(): string {
  const config: Config = {
    dictionaries: [colors, names],
    separator: "_",
    length: 2,
    style: "lowerCase",
  };

  const generatedName = uniqueNamesGenerator(config);

  const usernameWithNumber = `${generatedName}${Math.floor(Math.random() * 100)}`;

  return usernameWithNumber;
}
