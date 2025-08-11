# raas
Romance as a service - as a bot

A tiny Telegram bot you feed with **moments** and **details** about your partner. It stores them in a ChromaDB knowledge base, and (soon) at random intervals will pull a nugget, send it to an LLM, and craft sweet reminders, date ideas, or “don’t-forget” nudges. Think of it as a gentle, automated love wingman that remembers everything and prompts you at the right time. Power of tech, amiright?

---

## How it works (high level)

* You DM the bot things like `/add_detail She loves pistachio gelato` or `/add_moment Our first beach walk in Diani`.
* The bot saves each entry into **ChromaDB** (running locally via Docker).
* Later (to be implemented), a scheduler will:

  1. Fetch a random or context-matching detail/moment,
  2. Pass it to an LLM,
  3. Send you a charming suggestion or reminder in Telegram.

**Stack:** Go (Telegram bot) + ChromaDB (Docker image).

---

## Prerequisites

* A Telegram bot and **bot token** (create via `@BotFather`).
* Your **Telegram chat ID** so the bot talks ONLY to you.
* Docker and Docker Compose installed.
* Optional server (VPS) if you want it running 24/7.

---

## Configuration

Copy the provided `env.example` to `.env` and fill in values:

```dotenv
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
TENANT_NAME=
DATABASE_NAME=
```

* `TELEGRAM_BOT_TOKEN`: token from `@BotFather`
* `TELEGRAM_CHAT_ID`: your personal chat ID (the bot will gate all messages to this ID)
* `TENANT_NAME`, `DATABASE_NAME`: ChromaDB tenant/database to use (e.g. `raas` / `raas`)

---

## Commands (current)

* `/add_detail <text>` — saves a small fact or preference
  *Example:* `/add_detail She prefers window seats on flights`
* `/add_moment <text>` — saves a memory or event (stubbed for now; wiring follows)
* `<text>` — any other message triggers the default handler which gets relevant details and moments and (soon) will generate a response using an LLM.

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
   /add_detail Her favorite author is Chimamanda
   /add_moment The surprise dinner at her favorite restaurant
   ```
3. The bot confirms saves to Chroma.
4. (Upcoming) A scheduler periodically nudges you with ideas like:

   * “Pick up lilies on your way home btw.”
   * “Book a quiet table and bring that Chimamanda short story collection.”

---

## Notes and roadmap

* **Scheduling**: periodic prompt engine (cron-like or ticker-based) to be added.
* **LLM integration**: plug your preferred model/provider to generate messages. Or maybe try OpenRouter for once.
* **Search**: semantic retrieval for richer matching (e.g., “anniversary ideas”).
* **Export/backup**: snapshot the knowledge base.
* **Multi-user**: optional later; currently gated to a single `TELEGRAM_CHAT_ID`. But can be cute to have a couple's shared bot. Get what I mean ? 

---


## Troubleshooting

* **Bot doesn’t reply**: verify `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID`, ensure the bot is started and you’ve initiated a chat with it on Telegram.
* **ChromaDB not reachable**: confirm Docker is running, the container is healthy, and the bot can reach it (default `localhost:8000` if running on the same host).
* **Env not loaded**: ensure `.env` is present and your app loads it before starting.

---

Happy automating the romance ✨
