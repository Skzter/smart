import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";

let options = {
  // If true, debugging information is printed in the console.
  debug: true,
  // The OpenAI model (https://platform.openai.com/docs/models/overview)
  model: "gpt-5-mini-2025-08-07",
};

test("Suche nach Flügen und Anzeige von Flugrouten: Mallorca ab Leipzig-Halle, 3 Wochen ab 01.12.2025", async ({ page }) => {
  await page.goto("http://localhost:8082/");
  await auto("Gib als Reiseziel 'Mallorca' ein", { page, test }, options);
  await auto("Wähle als Abflughafen 'Leipzig-Halle' aus", { page, test }, options);
  await auto("Setze die Reisedauer auf 3 Wochen", { page, test }, options);
  await auto("Wähle als Startdatum den 01.12.2025", { page, test }, options);
  await auto("Drücke auf 'Suchen'", { page, test }, options);

  const offerCount = await auto("Prüfe ob mehr als 0 Flugangebote angezeigt werden", { page, test }, options);
  expect(offerCount).toBe(true)
});
