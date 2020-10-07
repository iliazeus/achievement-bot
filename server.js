import { env } from "process";

import capitalize from "capitalize";
import express from "express";
import FormData from "form-data";
import fetch from "node-fetch";
import jimp from "jimp";
import sharp from "sharp";
import winston from "winston";

// fixes https://github.com/oliver-moran/jimp/pull/951
// TODO: remove when merged
import configureJimp from "@jimp/custom";
import jimpPrintPlugin from "@jimp/plugin-print";
configureJimp({ plugins: [jimpPrintPlugin] }, jimp);

const logger = winston.createLogger({
  transports: [
    new winston.transports.Console({
      level: "info",
      format: winston.format.combine(
        winston.format.colorize(),
        winston.format.timestamp(),
        winston.format.printf(({ level, message, label, timestamp }) => {
          return `${timestamp} [${label ?? "global"}] ${level}: ${message}`;
        })
      ),
      handleExceptions: true,
    }),
    new winston.transports.File({
      level: "verbose",
      format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.printf(({ level, label, timestamp, ...rest }) => {
          return `${timestamp} [${label ?? "global"}] ${level}: ${JSON.stringify(rest)}`;
        })
      ),
      filename: "server.log",
      handleExceptions: true,
    }),
  ],
});

logger.info(`app starting`);

const USE_WEBHOOK = env.npm_package_config_useWebhook;

const TELEGRAM_TOKEN = env.npm_package_config_telegramToken;
const TELEGRAM_API_URL = `https://api.telegram.org/bot${TELEGRAM_TOKEN}`;

const TEMP_CHAT_ID = env.npm_package_config_tempChatId;

const WEBHOOK_ENDPOINT = env.npm_package_config_webhookEndpoint;
const WEBHOOK_PORT = Number.parseInt(env.npm_package_config_webhookPort);

const TEMPLATE_PATH = env.npm_package_config_templatePath;
const FONT_PATH = env.npm_package_config_fontPath;
const MAX_TEXT_WIDTH = Number.parseInt(env.npm_package_config_maxTextWidth);
const TEXT_X = Number.parseInt(env.npm_package_config_textX);

logger.info(`loading template file ${TEMPLATE_PATH}`);
const JIMP_TEMPLATE = await jimp.create(TEMPLATE_PATH);

logger.info(`loading font file ${FONT_PATH}`);
const JIMP_FONT = await jimp.loadFont(FONT_PATH);

const handleUpdates = async (updates) => {
  for (const update of updates) {
    const updateLogger = logger.child({ label: update.update_id });

    updateLogger.info(`handling update with id ${update.update_id}`);
    updateLogger.verbose(`update contents`, { update });

    const query = update.inline_query;
    if (!query) continue;

    updateLogger.info(`handling query with id ${query.id}`);

    updateLogger.info(`creating jimp image`);

    const jimpImage = JIMP_TEMPLATE.clone();

    const imageText = capitalize.words(query.query);
    const imageHeight = jimpImage.getHeight();

    const textHeight = jimp.measureTextHeight(JIMP_FONT, imageText, MAX_TEXT_WIDTH);

    const imageTextX = TEXT_X;
    const imageTextY = Math.floor((imageHeight - textHeight) / 2);

    jimpImage.print(JIMP_FONT, imageTextX, imageTextY, imageText, MAX_TEXT_WIDTH);

    updateLogger.info(`creating sharp image`);

    const sharpImage = sharp(jimpImage.bitmap.data, {
      raw: {
        width: jimpImage.bitmap.width,
        height: jimpImage.bitmap.height,
        channels: 4,
      },
    });

    updateLogger.info(`converting to webp`);

    const webpBuffer = await sharpImage.webp({ lossless: true }).toBuffer();

    updateLogger.info(`uploading to temp chat`);

    const tempChatRequest = new FormData();
    tempChatRequest.append("chat_id", TEMP_CHAT_ID);
    tempChatRequest.append("sticker", webpBuffer, { contentType: "image/webp", filename: `${query.id}.webp` });
    tempChatRequest.append("disable_notification", "true");

    updateLogger.verbose(`request to sendSticker`);

    const tempChatResponse = await (
      await fetch(`${TELEGRAM_API_URL}/sendSticker`, {
        method: "POST",
        body: tempChatRequest,
      })
    ).json();

    updateLogger.verbose(`response from sendSticker`, { response: tempChatResponse });

    if (!tempChatResponse.ok) throw new Error(tempChatResponse.error);

    const tempChatMessage = tempChatResponse.result;
    const tempStickerFileId = tempChatMessage.sticker.file_id;

    updateLogger.info(`sending inline query answer`);

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

    updateLogger.verbose(`answer contents`, { answer });

    await fetch(`${TELEGRAM_API_URL}/answerInlineQuery`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(answer),
    });

    setTimeout(async () => {
      updateLogger.info(`deleting temp chat message`);

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
    try {
      await handleUpdates(request.body);
      response.status(200);
    } catch (error) {
      logger.error(`error handling updates`, { error });
      response.status(500);
    }
  });

  app.listen(WEBHOOK_PORT);
} else {
  let offset = 0;

  const fetchUpdates = async () => {
    let response;

    response = await fetch(`${TELEGRAM_API_URL}/getUpdates?offset=${offset}`, { timeout: 0 });

    const json = await response.json();
    if (!json.ok) throw new Error(json.error);

    const updates = json.result;
    if (updates.length > 0) offset = updates[updates.length - 1].update_id + 1;

    return updates;
  };

  logger.info(`fetching and discarding unanswered updates`);

  try {
    await fetchUpdates();
  } catch (error) {
    logger.error(`error fetching updates`, { error });
  }

  while (true) {
    let updates;

    try {
      updates = await fetchUpdates();
    } catch (error) {
      logger.error(`error fetching updates`, { error });
      continue;
    }

    try {
      await handleUpdates(updates);
    } catch (error) {
      logger.error(`error handling updates`, { error });
      continue;
    }
  }
}
