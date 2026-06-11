package utils

import (
    "encoding/base64"
    "encoding/json"
    "net/http"
    "strings"
    "fmt"
    "context"
    "time"
    //"github.com/davecgh/go-spew/spew"
    //bin "github.com/gagliardetto/binary"
    "bytes"
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/portto/solana-go-sdk/client"
    "github.com/shopspring/decimal"
    //"log"
    "os"
    "github.com/fatih/color"
)

/** Type / Structure definition zone **/
type TransactionResponse struct {
  JSONRPC string `json:"jsonrpc"`
  Result  struct {
    BlockTime int64 `json:"blockTime"`
    Meta      struct {
      ComputeUnitsConsumed int64         `json:"computeUnitsConsumed"`
      Err                  interface{}   `json:"err"`
      Fee                  int64         `json:"fee"`
      InnerInstructions    []interface{} `json:"innerInstructions"`
      LoadedAddresses      struct {
        Readonly []interface{} `json:"readonly"`
        Writable []interface{} `json:"writable"`
      } `json:"loadedAddresses"`
      LogMessages       []string      `json:"logMessages"`
      PostBalances      []int64       `json:"postBalances"`
      PostTokenBalances []interface{} `json:"postTokenBalances"`
      PreBalances       []int64       `json:"preBalances"`
      PreTokenBalances  []interface{} `json:"preTokenBalances"`
      Rewards           []interface{} `json:"rewards"`
      Status            struct {
        Ok interface{} `json:"Ok"`
      } `json:"status"`
    } `json:"meta"`
    Slot        int64 `json:"slot"`
    Transaction struct {
      Message struct {
        AccountKeys []string `json:"accountKeys"`
        Header      struct {
          NumReadonlySignedAccounts   int64 `json:"numReadonlySignedAccounts"`
          NumReadonlyUnsignedAccounts int64 `json:"numReadonlyUnsignedAccounts"`
          NumRequiredSignatures       int64 `json:"numRequiredSignatures"`
        } `json:"header"`
        Instructions []struct {
          Accounts       []int64 `json:"accounts"`
          Data           string  `json:"data"`
          ProgramIDIndex int64   `json:"programIdIndex"`
        } `json:"instructions"`
        RecentBlockhash string `json:"recentBlockhash"`
      } `json:"message"`
      Signatures []string `json:"signatures"`
    } `json:"transaction"`
  } `json:"result"`
  ID int64 `json:"id"`
}

// Create a struct to represent the data
type Transaction struct {
  Address   string `json:"address"`
  Signature string `json:"signature"`
  Amount    int64  `json:"amount"`
}

// Create a struct to represent the data
type SolWallet struct {
  Address string  `json:"address"`
  Amount  float64 `json:"amount"`
}



/** Variable definition zone **/
// processNewSolDonation is a callback set from main.go
var processNewSolDonation func(addr, sig string, amount int64, memo string)

// Define a slice of Transaction objects
var transactions []Transaction
var solWallets = map[int]SolWallet{}

var firstRun bool = true

// Mainnet
var solClient = client.NewClient("https://api.mainnet-beta.solana.com")

// Prevent re-processing the same transaction on restart or duplicate polls
var processedSignatures = make(map[string]bool)

// Track tx's already processed per wallet address | Stop looping through every tx on server start
var lastProcessedSig = make(map[string]solana.Signature)

// Color definitions
//Green
var greenD = color.New(color.FgGreen).SprintFunc()
var green = color.New(color.FgHiGreen).SprintFunc()
var purple = color.New(color.FgHiMagenta).SprintFunc()
var yellow = color.New(color.FgHiYellow).SprintFunc()
// Keep color functions up with color var's
func greenTextDark(s string) string{
    return greenD(s)
}
func greenText(s string) string{
    return green(s)
}
func purpleText(s string) string{
    return purple(s)
}
func yellowText(s string) string{
    return yellow(s)
}

// persistLastSig writes the last processed signature for a wallet so we don't
// re-scan the same history after every server restart.
func persistLastSig(walletAddr string, sig solana.Signature) {
	data := map[string]string{
		walletAddr: sig.String(),
	}
	b, _ := json.MarshalIndent(data, "", "  ")
	_ = os.WriteFile("last_sigs.json", b, 0644)
}



// loadLastSigs reads previously saved signatures on startup.
func loadLastSigs() {
	b, err := os.ReadFile("last_sigs.json")
	if err != nil {
		return
	}
	var data map[string]string
	if json.Unmarshal(b, &data) == nil {
		for addr, sigStr := range data {
			if sig, err := solana.SignatureFromBase58(sigStr); err == nil {
				lastProcessedSig[addr] = sig
			}
		}
	}
}



func StartMonitoringSolana() {
    loadLastSigs()      // Restore cursor position and stop repeating previous tx's
    for {
        getTransactionsForAddresses()
    }
}



// CheckTransactionSolana checks for a matching Solana transaction and returns the full tx data if found
// Returns (found bool, txData interface{})
func CheckTransactionSolana(amt string, addr string, max_depth int) (bool, interface{}) {
	decAmountReceived, _ := decimal.NewFromString(amt)
	decMultiplier := decimal.NewFromFloat(1000000000)
	result := decAmountReceived.Mul(decMultiplier)
	amountSent := result.IntPart()

	fmt.Printf("🔍 [CheckTransactionSolana] Checking %s for %d lamports (requested: %s)\n", addr, amountSent, amt)

	startIndex := len(transactions) - max_depth
	if startIndex < 0 {
		startIndex = 0
	}

	for i := startIndex; i < len(transactions); i++ {
		transaction := transactions[i]
		if transaction.Address == addr && transaction.Amount == amountSent {
			fmt.Printf(" [CheckTransactionSolana] MATCH FOUND! Signature: %s\n", transaction.Signature)

			fullTx := fetchFullTransaction(transaction.Signature)
			if fullTx != nil {
				fmt.Println(" [CheckTransactionSolana] Full tx data fetched successfully")
			} else {
				fmt.Println(" [CheckTransactionSolana] Full tx fetch returned nil")
			}
			return true, fullTx
		}
	}

	fmt.Printf(" [CheckTransactionSolana] No match found for %s / %d lamports in last %d txs\n", addr, amountSent, max_depth)
	return false, nil
}
// End of CheckTransactionSolana


func SetSolWallets(sW map[int]SolWallet) {
  solWallets = sW
}

// SetSolanaDonationCallback registers the callback from main.go
// so new incoming SOL transactions with memos trigger the alert + TTS
func SetSolanaDonationCallback(fn func(addr, sig string, amount int64, memo string)) {
    fmt.Println(">>> [DEBUG] SetSolanaDonationCallback was called")
    processNewSolDonation = fn
}

// getTransactionsForAddresses polls for new Solana transactions.
// It always fetches a small recent page (newest first) and stops
func getTransactionsForAddresses() {
	for _, wallet := range solWallets {
		sameBalance := false
		wallet, sameBalance = checkSameBalanceSol(wallet)

		if sameBalance {
			// Debug line
            //fmt.Println("Sol wallet the same balance, not getting new txs")
			fmt.Println("Awaiting first donation...")
            time.Sleep(10 * time.Second)
            fmt.Println("Checking wallet...")
			continue
		}

		//fmt.Println("Sol wallet changed balance → fetching NEW txs only")
        fmt.Println("Waiting for donations...")

		endpoint := rpc.MainNetBeta_RPC
		client := rpc.New(endpoint)

		// Always fetch a small page of the most recent signatures (newest first).
		limit := 20
		opts := &rpc.GetSignaturesForAddressOpts{
			Limit: &limit,
		}

		out, err := client.GetSignaturesForAddressWithOpts(
			context.TODO(),
			solana.MustPublicKeyFromBase58(wallet.Address),
			opts,
		)
		if err != nil {
			fmt.Printf("RPC GetSignatures error: %v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, sigInfo := range out {
			sigStr := sigInfo.Signature.String()

			// If we have already processed this signature, stop.
			if processedSignatures[sigStr] {
				break
			}

			tAmount, newTrans := getTransactionAmount(sigStr, wallet.Address)
			if newTrans {
				addSolanaTransaction(wallet.Address, sigStr, tAmount)
			} else {
				fmt.Printf("SOL: No new tx for %s...\n", wallet.Address[:7])
			}

			// Update cursor (test keeping lastProcessedSig for the JSON file)
			lastProcessedSig[wallet.Address] = sigInfo.Signature
			persistLastSig(wallet.Address, sigInfo.Signature)

			time.Sleep(6 * time.Second)
		}

		time.Sleep(5 * time.Second)
	}
}
// End of getTransactionsForAddresses


// addSolanaTransaction is called when a new incoming SOL tx is detected.
// It extracts the memo and triggers the alert/TTS callback.
// The double processedSignatures guard prevents duplicate success logs.
func addSolanaTransaction(addr, sig string, amount int64) {
	// First guard — exit immediately if we have already seen this signature
	if processedSignatures[sig] {
		return
	}
	processedSignatures[sig] = true

	transaction := Transaction{
		Address:   addr,
		Signature: sig,
		Amount:    amount,
	}

	if amount <= 50000 { // prevent dust/spam
		return
	}

	// Fetch full tx (jsonParsed) so ExtractSolanaMemo has good data
	fullTx := fetchFullTransaction(sig)
	memo := ExtractSolanaMemo(fullTx)
	if memo == "" {
		memo = "Anonymous Donation"
	}

	// Print success message with memo for nice console output
    fmt.Printf("%s %s%.6f SOL || %s%s\n",
        greenTextDark("[DONATION]"),
        yellowText("Amount: "),
        float64(amount)/1e9,
        purpleText("Message: "),
        memo)

	// Second guard — belt-and-suspenders in case of any re-entrancy or timing
	if processedSignatures[sig] {
		return
	}

    processNewSolDonation(addr, sig, amount, memo)

	// Trigger the actual alert + TTS
	if processNewSolDonation != nil {
		processNewSolDonation(addr, sig, amount, memo)
	}

	transactions = append(transactions, transaction)
}
// End of addSolanaTransaction

func CreatePendingSolDono(name string, message string, mediaURL string, amountNeeded float64) SuperChat {
    pendingDono := SuperChat{
        Name:         name,
        Message:      message,
        MediaURL:     mediaURL,
        AmountNeeded: amountNeeded,
        Completed:    false,
        CreatedAt:    time.Now().String(),
        CheckedAt:    time.Now().String(),
        CryptoCode:   "SOL",
      }
      return pendingDono
}

func containsTransaction(sig string) bool {
    // searches in reverse order in order to search newest transactions first to avoid needless loops
    for i := len(transactions) - 1; i >= 0; i-- {
        if transactions[i].Signature == sig {
          return true
        }
    }
    return false
}

func checkSameBalanceSol(wallet SolWallet) (SolWallet, bool) {
    amt, _ := getSOLBalance(wallet.Address)
      if amt == wallet.Amount {
        return wallet, true
      } else {
        wallet.Amount = amt
        return wallet, false
      }

}

func getSOLBalance(address string) (float64, error) {

    if address == "" {
        return 0, nil
    }
    balance, err := solClient.GetBalance(
    context.TODO(), // request context
    address,        // wallet to fetch balance for
      )
    if err != nil {
        return 0, err
    }
        return float64(balance) / 1e9, nil
}

// getTransactionAmount safely parses a Solana transaction and extracts amount + memo
func getTransactionAmount(sig, addr string) (int64, bool) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Recovered from panic in getTransactionAmount on %s: %v\n", sig[:12]+"...", r)
            time.Sleep(5 * time.Second)
        }
    }()

    if containsTransaction(sig) {
        return 0, false
    }

    url := "https://api.mainnet-beta.solana.com"
    requestBody := fmt.Sprintf(`{
        "jsonrpc": "2.0",
        "id": 1,
        "method": "getTransaction",
        "params": [
            "%s",
            "json"
        ]
    }`, sig)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
    if err != nil {
        fmt.Println("Error creating HTTP request:", err)
        return 0, false
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 15 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error sending HTTP request:", err)
        return 0, false
    }
    defer resp.Body.Close()

    var responseBody bytes.Buffer
    _, err = responseBody.ReadFrom(resp.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        return 0, false
    }

    var tr TransactionResponse
    err = json.Unmarshal(responseBody.Bytes(), &tr)
    if err != nil {
        fmt.Printf("[!ERROR!] JSON unmarshal failed for %s: %v\n", sig[:12]+"...", err)
        return 0, false
    }

    // === SAFE INDEX CHECKS (prevents the panic) ===
    if len(tr.Result.Meta.PreBalances) == 0 ||
        len(tr.Result.Meta.PostBalances) == 0 ||
        len(tr.Result.Transaction.Message.AccountKeys) == 0 {
        return 0, false
    }

    initialAmount := tr.Result.Meta.PreBalances[0]
    endingAmount := tr.Result.Meta.PostBalances[0]
    fromAddr := tr.Result.Transaction.Message.AccountKeys[0]
    fee := tr.Result.Meta.Fee

    endingPlusFee := endingAmount + fee
    amountSent := initialAmount - endingPlusFee

    if fromAddr == addr {
        amountSent *= -1
    }

   // Only count positive incoming amounts as potential donations
    if amountSent <= 0 {
        return 0, false
    }

        // clean up duplicate messages by commenting this line
        //fmt.Printf("✅ SOL: %s... Received: %d lamports (%.6f SOL)\n",
        //	addr[:5], amountSent, float64(amountSent)/1e9)

    return amountSent, true
}
    // getTransactionAmount END


    func printSolTx(fromAddr, checkAddr, toAddr string, amountSent int64, sig string) {

      decAmountSent := decimal.NewFromInt(amountSent)
      decMultiplier := decimal.NewFromFloat(0.000000001)
      amt := decAmountSent.Mul(decMultiplier)

      if fromAddr == checkAddr {
        fmt.Println("\nTRANSACTION OUT:")
      } else {
        fmt.Println("\nTRANSACTION IN:")
      }
      fmt.Println("To:", toAddr[:7])
      fmt.Println("Sent:", amt)
      fmt.Println("sig:", sig[:7])
    }

    // fetchFullTransaction gets the full parsed tx.
    // It stays completely silent during normal retries and only logs on final success or total failure.
    func fetchFullTransaction(signature string) interface{} {
        url := "https://api.mainnet-beta.solana.com"

        for attempt := 1; attempt <= 12; attempt++ {
            requestBody := fmt.Sprintf(`{
                "jsonrpc": "2.0",
                "id": 1,
                "method": "getTransaction",
                "params": [
                    "%s",
                    {
                        "encoding": "jsonParsed",
                        "maxSupportedTransactionVersion": 0,
                        "commitment": "finalized"
                    }
                ]
            }`, signature)

            req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
            if err != nil {
                time.Sleep(time.Duration(attempt*200) * time.Millisecond)
                continue
            }
            req.Header.Set("Content-Type", "application/json")

            client := &http.Client{Timeout: 15 * time.Second}
            resp, err := client.Do(req)
            if err != nil {
                time.Sleep(time.Duration(attempt*500) * time.Millisecond)
                continue
            }
            defer resp.Body.Close()

            var txResponse map[string]interface{}
            if err := json.NewDecoder(resp.Body).Decode(&txResponse); err != nil {
                time.Sleep(time.Duration(attempt*400) * time.Millisecond)
                continue
            }

            if result, exists := txResponse["result"]; exists && result != nil {
                // Debug line for success
                //fmt.Printf(" [RPC] Full tx data received for %s\n", signature[:12]+"...")
                // Production line
                fmt.Printf("%s %s\n",greenText("[DONATION RECIEVED]"), signature[:24] + "...")
                return result
            }

            // Completely silent during normal "not yet available" retries
		time.Sleep(time.Duration(attempt*800) * time.Millisecond)
	}

	fmt.Printf("[FAIL][RPC] Failed to fetch tx %s after 12 attempts\n", signature[:12]+"...")
	return nil
}
// End of fetchFUllTransaction


// fetchPlainTransaction is a lightweight fallback used only when we need the original base64 memo shape.
func fetchPlainTransaction(signature string) interface{} {
	url := "https://api.mainnet-beta.solana.com"
	requestBody := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "getTransaction",
		"params": ["%s", "json"]
	}`, signature)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var txResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&txResponse)

	if result, ok := txResponse["result"]; ok && result != nil {
		return result
	}
	return nil
}

// ExtractSolanaMemo extracts the memo from a full Solana getTransaction response.
// Handles Phantom, Ledger, CLI, etc. via instructions + logMessages fallback.
func ExtractSolanaMemo(tx interface{}) string {
	if tx == nil {
		return ""
	}

	memo := ""

	if txMap, ok := tx.(map[string]interface{}); ok {
		// Primary path: transaction.message.instructions
		var instructions []interface{}
		if transaction, ok := txMap["transaction"].(map[string]interface{}); ok {
			if message, ok := transaction["message"].(map[string]interface{}); ok {
				if insts, ok := message["instructions"].([]interface{}); ok {
					instructions = insts
				}
			}
		}

		// Scan for Memo program
		for _, inst := range instructions {
			if instMap, ok := inst.(map[string]interface{}); ok {
				programID := ""
				if pid, ok := instMap["programId"].(string); ok {
					programID = pid
				} else if pid, ok := instMap["programID"].(string); ok {
					programID = pid
				}

				if strings.Contains(programID, "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr") ||
				   strings.Contains(programID, "Memo1Uh8x") {
					if data, ok := instMap["data"].(string); ok {
						decoded, err := base64.StdEncoding.DecodeString(data)
						if err == nil {
							memo = string(decoded)
						} else {
							memo = data
						}
						break
					}
				}
			}
		}

		// Fallback for Ledger / wallets that only log the memo
		if memo == "" {
			if meta, ok := txMap["meta"].(map[string]interface{}); ok {
				if logMessages, ok := meta["logMessages"].([]interface{}); ok {
					for _, logEntry := range logMessages {
						if logStr, ok := logEntry.(string); ok {
							if strings.Contains(logStr, "Memo (len") {
								if idx := strings.LastIndex(logStr, `"`); idx > 0 {
									start := strings.LastIndex(logStr[:idx], `"`)
									if start != -1 {
										memo = strings.TrimSpace(logStr[start+1 : idx])
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}

	memo = strings.TrimSpace(memo)
	if len(memo) > 280 {
		memo = memo[:280]
	}
	return memo
}
// End of ExtractSolanaMemo

// Helper to print map keys for debugging (add this too)
func getMapKeys(m map[string]interface{}) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
