import { chromium } from 'playwright';
import { GeminiClient } from './llms/gemini-client';
import * as dotenv from 'dotenv';

dotenv.config();

async function main() {
  const gemini = new GeminiClient();
  const prompt = 'Give me a short summary of the importance of AI in 2025.';

  try {
    const result = await gemini.generateContent(prompt);
    console.log('Gemini response:', result);
  } catch (error) {
    console.error('Error using GeminiClient:', error);
  }
}

main();
