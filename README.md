# 88Pay

- Self-hosted, noncustodial crypto-currency (currently Monero(XMR), Ethereum(ETH), Solana(SOL), PAINT, HEX, MATIC, BUSD, SHIBA_INU, USDC, TETHER, WBTC, and PNK ) superchat system written in Go.
- [SOON] Reads TTS messages from Solana transaction memo fields for viewer interaction.
- [SOON] Word filtering for TTS
- Provides notifications and a progress bar usable in OBS as well as admin pages for settings like minimum donos.
- Settings pages /user /userobs (default login is user:admin password:hunter123)

# Current Goals
[ ] - Add memo field parsing for Solana transaction for custom TTS from viewers
[ ] - Remove all ferret.cash references
[ ] - General UI improvements
[ ] - Add "88Pay" logos
[ ] - Add TTS word filtering to stop undesireable donations

# [SOON] Windows Installation
1. Download/Clone 88pay into a folder of your choice
2. Install GoLang and TDM-GCC for windows
3. Run "88pay-start.bat"
4. Close the terminal window when you want to stop recieving donations

# Linux/Mac Installation

1. ```apt install golang```
2. ```git clone https://github.com/pautown/paulpay.git```
3. ```cd shadowchat```
4. ```go install github.com/skip2/go-qrcode@latest```
5. ```go run main.go```



A webserver at 127.0.0.1:8900 is running.

# Features
- Youtube Media 
- Sound and GIF for donos
- TTS integration for donos
- 9 cryptos supported (XMR, SOLANA, ETH, and six ERC-20 tokens)
- Keeping track of USD value
- Selection of which dono methods are available

This is currently designed to be run on a cloud server with nginx proxypass for TLS.

# Monero Wallet Setup

1. Generate a view only wallet using the `monero-wallet-gui` from getmonero.org. Preferably with no password (need to change code if you have a password)
2. Upload the newly generated `walletname_viewonly` and `walletname_viewonly.keys` files in the user account.
3. Download the `monero-wallet-rpc` binary that is bundled with the getmonero.org wallets.
4. Place the 'monero-wallet-rpc' inside monero folder

# Usage
- Visit 127.0.0.1:8900/user to view your user settings
- Visit 127.0.0.1:8900/userobs to view your user OBS settings
- Visit 127.0.0.1:8900/alert to see notifications (only have one of these open at a time, preferrably in the OBS screen)
- Visit 127.0.0.1:8900/progressbar to see the OBS progressbar which gets modified in the OBS settings url
- The default username is `admin` and password `hunter123`. Change these in the http://127.0.0.1:8900/user panel

## License

**88pay** is licensed under the **GNU General Public License v3.0** (GPLv3).

See the [LICENSE](LICENSE) file for the full license text.

### Credits & Origin

- **88pay** is a fork and active development of [paulpay](https://github.com/pautown/paulpay) by **Paul Town**.
- The base logic originally comes from [shadowchat](https://git.sr.ht/~anon_/shadowchat) (thank you to the anonymous author — without that foundation this project wouldn't exist).
- Major rewrites, Solana support, memo parsing, Windows workflow, and streamer-focused features added by **Zennox-97**.

Thank you to everyone who contributed to the original code!

### Donate

To support Paul who this fork would not be possible without, send XMR to his wallet here:
`88K988HXHBTZZEFACejzJRDe7zMiKviesFKWtq4Q3Bo6VZfPZDWFzbod4Kn7SudVSBKhu5GqMUqBUXFNj5wBLyWuNWe4nqN`

To support me (Zennox) please send Solana to this address
`5Ci84K1CJRVyWJHkxhsNuscBsXQZN7zG5AuVqWYYnjtf`
