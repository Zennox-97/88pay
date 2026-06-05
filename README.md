# 88 Pay

Free, self-hosted cryptocurrency donation service that interfaces with OBS.

# What is 88 Pay?
88 Pay is a self-hosted streaming donation handler that utilizes cryptocurrency and is impossible to debank.
After Entropy got debanked many streamers had to scramble to set up a powerchat, which itself will ban you if it gets
too many reports of "hateful" streams.
88 Pay casts no judgement and is being built because a truly independent system for getting paid by viewers for streams needs to exist.
Solana is the primary coin for donation. The memo field in each transaction works perfectly as a custom TTS message input.
Currently the project also works with Ethereum, Monero and a handful of others.

Stream, say, and do whatever you want and get paid to do it. That's the point of this project.


## Current Goals
[X] - Add memo field parsing for Solana transaction for custom TTS from viewers
[X] - Remove all ferret.cash references
[ ] - General UI improvements
[ ] - Add "88Pay" logos
[ ] - Add TTS word filtering to stop undesireable donations
- [SOON] Reads TTS messages from Solana transaction memo fields for viewer interaction.
- [SOON] Word filtering for TTS
- Provides notifications and a progress bar usable in OBS as well as admin pages for settings like minimum donos.

## [SOON] Windows Installation

1. Download/Clone 88pay into a folder of your choice
2. Install GoLang 1.25 and TDM-GCC for windows
-   https://github.com/jmeubank/tdm-gcc/releases/download/v10.3.0-tdm64-2/tdm64-gcc-10.3.0-2.exe
-   https://go.dev/dl/go1.25.11.windows-amd64.msi
3. Run "8start.bat"
4. Close the terminal window when you want to stop recieving donations

## Linux/Mac Installation

1. ```apt install golang```
2. ```git clone https://github.com/Zennox-97/88pay.git```
5. ```./start.sh```

## How to use 88 Pay
A webserver at 127.0.0.1:8900 is running.
Open your web browser of choice and in the top URL bar type `127.0.0.1:8900`
You will see a login page, the login is 
Username: admin
Password: hunter123
CHANGE THE PASSWORD IMMEDIATELY
If the login doesnt work, close the server and open it again. Sometimes it needs to be launched twice to set up the databse
Once logged in, go to the settings tabs and set up your crypto wallet address, custom images, sounds and donation tiers.
Interface the OBS links for alerts and donation bar directly into an OBS browser element
Start streaming and enjoy receiving 100% of your donations (minus the super tiny transaction fees)
88 Pay takes 0% of your donations. You keep everything.

# Features
- Sound and GIF for donos
- TTS integration for donos
- Keeping track of USD value
- Selection of which dono methods are available

This is currently designed to be run on a cloud server with nginx proxypass for TLS.

# Monero Wallet Setup

1. Generate a view only wallet using the `monero-wallet-gui` from getmonero.org. Preferably with no password (need to change code if you have a password)
2. Upload the newly generated `walletname_viewonly` and `walletname_viewonly.keys` files in the user account.
3. Download the `monero-wallet-rpc` binary that is bundled with the getmonero.org wallets.
4. Place the 'monero-wallet-rpc' inside monero folder

# Directories for the webserver
IMMEDIATELY CHANGE THE ADMIN PASSWORD
- Visit 127.0.0.1:8900/user to view your user settings
- Visit 127.0.0.1:8900/userobs to view your user OBS settings
- Visit 127.0.0.1:8900/alert to see notifications (only have one of these open at a time, preferrably in the OBS screen)
- Visit 127.0.0.1:8900/progressbar to see the OBS progressbar which gets modified in the OBS settings url
- The default username is `admin` and password `hunter123`. Change these in the http://127.0.0.1:8900/user panel

# License
GPLv3

### Origin
The original fork has this text

"This comes from [https://git.sr.ht/~anon_/shadowchat](https://git.sr.ht/~anon_/shadowchat) and the base logic (mostly rewritten now) is not Paul's original
work, although without the base logic I would have never started doing this, so thank you to the great mind behind this."

This is a fork from [fork source here]. Would not have been possible without Paul's work, which is now being modified by me.

### Donate

To support the creator that made this fork possible, send XMR to Paul Town at:
`88K988HXHBTZZEFACejzJRDe7zMiKviesFKWtq4Q3Bo6VZfPZDWFzbod4Kn7SudVSBKhu5GqMUqBUXFNj5wBLyWuNWe4nqN`

To support the creator of 88 Pay, send Solana to me at:
`5Ci84K1CJRVyWJHkxhsNuscBsXQZN7zG5AuVqWYYnjtf`
