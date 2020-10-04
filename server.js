import { env } from "process";

import capitalize from "capitalize";
import express from "express";
import FormData from "form-data";
import fetch from "node-fetch";
import jimp from "jimp";
import sharp from "sharp";

const USE_WEBHOOK = env.npm_package_config_useWebhook;

const TELEGRAM_TOKEN = env.npm_package_config_telegramToken;
const TELEGRAM_API_URL = `https://api.telegram.org/bot${TELEGRAM_TOKEN}`;

const TEMP_CHAT_ID = env.npm_package_config_tempChatId;

const WEBHOOK_ENDPOINT = env.npm_package_config_webhookEndpoint;
const WEBHOOK_PORT = env.npm_package_config_webhookPort;

const TEMPLATE_PATH = env.npm_package_config_templatePath;
const FONT_PATH = env.npm_package_config_fontPath;

const JIMP_TEMPLATE = await jimp.create(TEMPLATE_PATH);
const JIMP_FONT = await jimp.loadFont(FONT_PATH);

const handleUpdates = async (updates) => {
  for (const update of updates) {
    const query = update.inline_query;
    if (!query) continue;

    const jimpImage = JIMP_TEMPLATE.clone();

    jimpImage.print(JIMP_FONT, 200, 50, capitalize.words(query.query));

    const sharpImage = sharp(jimpImage.bitmap.data, {
      raw: {
        width: jimpImage.bitmap.width,
        height: jimpImage.bitmap.height,
        channels: 4,
      },
    });

    const webpBuffer = await sharpImage.webp({ lossless: true }).toBuffer();

    const tempChatRequest = new FormData();
    tempChatRequest.append("chat_id", TEMP_CHAT_ID);
    tempChatRequest.append("document", webpBuffer, { contentType: "image/webp", filename: `${query.id}.webp` });

    const tempChatResponse = await (
      await fetch(`${TELEGRAM_API_URL}/sendDocument`, {
        method: "POST",
        body: tempChatRequest,
      })
    ).json();

    const tempChatMessage = tempChatResponse.result;
    const tempStickerFileId = tempChatMessage.sticker.file_id;

    const answer = {
      inline_query_id: query.id,
      results: [
        {
          type: "sticker",
          id: query.id,
          sticker_file_id: tempStickerFileId,
        },
      ],
    };

    await fetch(`${TELEGRAM_API_URL}/answerInlineQuery`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(answer),
    });

    setTimeout(async () => {
      await fetch(`${TELEGRAM_API_URL}/deleteMessage`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ chat_id: TEMP_CHAT_ID, message_id: tempChatMessage.message_id }),
      });
    }, 1 * 1000);
  }
};

if (USE_WEBHOOK) {
  const app = express();

  app.use(express.json());

  app.post(WEBHOOK_ENDPOINT, async (request, response) => {
    await handleUpdates(request.body);
    response.status(200);
  });

  app.listen(WEBHOOK_PORT);
} else {
  let offset = 0;

  const fetchUpdates = async () => {
    let response;

    try {
      response = await fetch(`${TELEGRAM_API_URL}/getUpdates?offset=${offset}`, { timeout: 0 });
    } catch (error) {
      return [];
    }

    const json = await response.json();
    if (!json.ok) {
      throw new Error(json.error);
    }

    const updates = json.result;
    if (updates.length > 0) offset = updates[updates.length - 1].update_id + 1;

    return updates;
  };

  await fetchUpdates();

  while (true) {
    const updates = await fetchUpdates();
    await handleUpdates(updates);
  }
}
