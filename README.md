# raas
Romance as a service - as a bot

A tiny Telegram bot you feed with **moments** and **details** about your partner. It stores them in a ChromaDB knowledge base, and when you text it and at random intervals will pull a nugget, send it to an LLM, and craft sweet reminders, date ideas, or “don’t-forget” nudges. Think of it as a gentle, automated love wingman that remembers everything and prompts you at the right time. Power of tech, amiright?

---

## How it works (high level)

* You DM the bot things like `/add_detail She loves pistachio gelato` or `/add_moment Our first beach walk in Tasmania`.
* The bot saves each entry into **ChromaDB** (running locally via Docker).
* Later (to be implemented), a scheduler will:

  1. Fetch a random or context-matching detail/moment,
  2. Pass it to an LLM,
  3. Send you a charming suggestion or reminder in Telegram.
* The above can also be triggered when you just send a text to the bot

**Stack:** Go (Telegram bot) + ChromaDB (Docker image).

---

## Prerequisites

* A Telegram bot and **bot token** (create via `@BotFather`).
* Your **Telegram chat ID** so the bot talks ONLY to you.
* OPEN AI api key
* Docker and Docker Compose installed.
* Optional server (VPS) if you want it running 24/7.

---

## Configuration

Copy the provided `env.example` to `.env` and fill in values:

```dotenv
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
OPENAI_API_KEY=
TENANT_NAME=
DATABASE_NAME=
```

* `TELEGRAM_BOT_TOKEN`: token from `@BotFather`
* `TELEGRAM_CHAT_ID`: your personal chat ID (the bot will gate all messages to this ID)
* `OPENAI_API_KEY`: your personal Open AI Key. GPT 4.1 Nano is used so make sure you have access
* `TENANT_NAME`, `DATABASE_NAME`: ChromaDB tenant/database to use (e.g. `raas` / `raas`)

---

## Commands (current)

* `/add_detail <text>` — saves a small fact or preference
  *Example:* `/add_detail She prefers window seats on flights`
* `/add_moment <text>` — saves a memory or event
  *Example:* `/add_moment That look across the room at the museum when we both knew we hated it there`
* `<text>` — any other message triggers the default handler which gets relevant details and moments and  will generate a response using an LLM.

---

## Setup and Run

1. **Start ChromaDB (Docker)**

   ```bash
   docker compose up -d
   ```

2. **Build the Go binary**

   ```bash
   go build -o raas .
   ```

3. **Run the bot**

   ```bash
   ./raas
   ```

   Keep the process running (e.g., systemd, tmux, or a process manager) if deploying to a server.

---

## Typical flow

1. Start the bot.
2. DM your bot on Telegram:

   ```
   /add_detail She loves lilies and pistachio gelato
   /add_moment The surprise dinner at her favorite restaurant
   ```
   or send a normal text to get a curated response.
   
3. The bot confirms saves to Chroma.
4. (Upcoming) A scheduler periodically nudges you with ideas like:

   * “Pick up lilies on your way home btw.”
   * “Book a quiet table and bring that Chimamanda short story collection.”

---


## Troubleshooting

* Check the logs to see what's up my guy! 

---

Happy automating the romance ✨
