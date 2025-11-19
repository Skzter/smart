import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";

let options = {
  // If true, debugging information is printed in the console.
  debug: true,
  // The OpenAI model (https://platform.openai.com/docs/models/overview)
  model: "gpt-5-mini-2025-08-07",
};

test("Login: Nutzer meldet sich erfolgreich an", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto(
    "Warte auf die CookieAnfrage, klicke auf den blauen button 'Geht klar'",
    { page, test },
    options,
  );
  await auto(
    "Warte auf das Cashback Popup, klicke auf das Kreuz oben rechts",
    { page, test },
    options,
  );
  await auto("Klicke auf 'Anmelden'", { page, test });
  await auto(
    "Gib in das Feld mit dem Platzhalter 'E-Mail-Adresse' den Wert 'luca@auto.com'",
    { page, test },
    options,
  );
  await auto("Klicke auf 'weiter'", { page, test });

  const registrierung = await auto(
    "Ist 'Kundenkonto anlegen und Vorteile nutzen!' auf der seite zu sehen?",
    {
      page,
      test,
    },
    options,
  );
  expect(registrierung).toBe(true);
});
