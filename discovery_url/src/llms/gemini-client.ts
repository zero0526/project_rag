// src/llms/gemini-client.ts
import axios from 'axios';
import * as dotenv from 'dotenv';

dotenv.config();

const API_KEY = process.env.GEMINI_API_KEY;
const API_URL_TEMPLATE = `https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=${API_KEY}`;

interface GeminiResponse {
  candidates: Array<{
    content: {
      parts: Array<{
        text: string;
      }>;
    };
  }>;
}

export class GeminiClient {
  private apiUrl: string;

  constructor() {
    if (!API_KEY) throw new Error('GEMINI_API_KEY not set in environment');
    this.apiUrl = API_URL_TEMPLATE;
  }

  async generateContent(prompt: string): Promise<string> {
    const payload = {
      contents: [
        {
          parts: [
            { text: prompt }
          ]
        }
      ]
    };

    try {
      const response = await axios.post<GeminiResponse>(this.apiUrl, payload, {
        headers: {
          'Content-Type': 'application/json'
        }
      });

      return response.data.candidates?.[0]?.content?.parts?.[0]?.text ?? '';
    } catch (error: any) {
      console.error('Gemini API error:', error.response?.data || error.message);
      throw new Error('Failed to generate content');
    }
  }
}