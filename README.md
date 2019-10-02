# StupidUptimeBot

Restarts your Hetzner server via robot API if downtime is detected. 

# How is downtime detected?

The Hetzner server has a crontab to curl a URL every X minutes. If this URL is not curl'd within a certain timeframe, StupidUptimeBot issues a restart command to robot.

# Example Config

```json
{
  "Port": 8080,
  "Password": "lions-are-not-a-password",
  "BotToken": "123456789:AAAAAAAAAA-_eeeeeeeeeeee",
  "AlertUser": "@YourUser",
  "AlertChatID": 1234567,
  "Minutes": 10,
  "AutoRestartMultiplier": 3,
  "AdminUserID": 123456789,
  "HetznerUser": "K1337",
  "HetznerPassword": "neither-are-penguins",
  "HetznerIP": "255.255.255.255"
}
```
