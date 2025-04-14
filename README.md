# [blazium.app](https://blazium.app)

The primary source code for the Blazium website.

Uses [discord-inviter](https://github.com/blazium-engine/discord-inviter) to redirect [chat.blazium.app](https://chat.blazium.app) to a Blazium Discord server invite.

## Tip for local development

Use the `--local` flag to connect to `blazium.app` to not do failing requests to cerbero.
```
go run . --local
```