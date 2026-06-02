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
)

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


func StartMonitoringSolana() {
  for {
    getTransactionsForAddresses()
  }
}

// New CheckTransactionSolana

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
	processNewSolDonation = fn
}

/* pre-loopfix getTransactionsForAddresses
func getTransactionsForAddresses() {
  for _, wallet := range solWallets {
    sameBalance := false
    wallet, sameBalance = checkSameBalanceSol(wallet)

    if sameBalance {
      fmt.Println("Sol wallet the same balance, not getting new txs")
      time.Sleep(10 * time.Second)
    } else {
      fmt.Println("Sol wallet not the same balance, getting new txs")
      endpoint := rpc.MainNetBeta_RPC
      client := rpc.New(endpoint)
      out, err := client.GetSignaturesForAddress(
        context.TODO(),
        solana.MustPublicKeyFromBase58(wallet.Address),
      )
      if err != nil {
        panic(err)
      }
      for _, sig := range out {
        tAmount, newTrans := getTransactionAmount(sig.Signature.String(), wallet.Address)
        if newTrans {
          addSolanaTransaction(wallet.Address, sig.Signature.String(), tAmount)
        } else {
          fmt.Println("SOL: No new", wallet.Address[:7]+"... txs.")
        }

        time.Sleep(6 * time.Second)
      }
      time.Sleep(5 * time.Second)
    }
  }

} */

// New getTransactionsForAddresses (fixed tx looping)
func getTransactionsForAddresses() {
	for _, wallet := range solWallets {
		sameBalance := false
		wallet, sameBalance = checkSameBalanceSol(wallet)

		if sameBalance {
			fmt.Println("Sol wallet the same balance, not getting new txs")
			time.Sleep(10 * time.Second)
			continue
		}

		fmt.Println("Sol wallet changed balance → fetching NEW txs only")

		endpoint := rpc.MainNetBeta_RPC
		client := rpc.New(endpoint)

		// NEW: Only fetch signatures newer than the last one we processed
		var before solana.Signature
		if sig, ok := lastProcessedSig[wallet.Address]; ok {
			before = sig
		}

		limit := 20 // small batch = plenty for real-time watching
		opts := &rpc.GetSignaturesForAddressOpts{
			Limit:  &limit, // must be pointer
			Before: before, // solana.Signature type (not string)
		}

		out, err := client.GetSignaturesForAddressWithOpts(
			context.TODO(),
			solana.MustPublicKeyFromBase58(wallet.Address),
			opts,
		)
		if err != nil {
			fmt.Printf("❌ RPC GetSignatures error: %v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, sigInfo := range out {
			// Update our cursor immediately (this is the "remember where we left off")
			lastProcessedSig[wallet.Address] = sigInfo.Signature

			sigStr := sigInfo.Signature.String()

			tAmount, newTrans := getTransactionAmount(sigStr, wallet.Address)
			if newTrans {
				addSolanaTransaction(wallet.Address, sigStr, tAmount)
			} else {
				fmt.Printf("SOL: No new tx for %s...\n", wallet.Address[:7])
			}

			time.Sleep(6 * time.Second) // be nice to the RPC
		}

		time.Sleep(5 * time.Second)
	}
}



// addSolanaTransaction is called when a new incoming SOL tx is detected
func addSolanaTransaction(addr, sig string, amount int64) {
	// === DEDUPLICATION FIX ===
	if processedSignatures[sig] {
		return // already processed this tx
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

	fmt.Printf("SOL: %s... Received: %d lamports (%.6f SOL)\n", addr[:5], amount, float64(amount)/1e9)

	// Fetch memo
	fullTx := fetchFullTransaction(sig)
	memo := ExtractSolanaMemo(fullTx)

	// Trigger alert with memo
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

func getTransactionAmount(sig, addr string) (int64, bool) {
  defer func() {
    if r := recover(); r != nil {
      fmt.Println("Recovered from panic:", r)
      fmt.Println("Sleeping 10 seconds.")
      time.Sleep(10 * time.Second)
    }
  }()

  if !containsTransaction(sig) {
    url := "https://api.mainnet-beta.solana.com"
    requestBody := fmt.Sprintf(`
  {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "getTransaction",
    "params": [
      "%s",
      "json"
    ]
  }`, sig)

    // Create an HTTP POST request with the request body
    req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
    if err != nil {
      fmt.Println("Error creating HTTP request:", err)
      return 0, false
    }

    // Set the request header
    req.Header.Set("Content-Type", "application/json")

    // Send the HTTP request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
      fmt.Println("Error sending HTTP request:", err)
      return 0, false
    }
    defer resp.Body.Close()

    // Read the response body
    var responseBody bytes.Buffer
    _, err = responseBody.ReadFrom(resp.Body)
    if err != nil {
      fmt.Println("Error reading response body:", err)
      return 0, false
    }

    // Parse the response into a TransactionResponse struct
    var tr TransactionResponse
    err = json.Unmarshal(responseBody.Bytes(), &tr)
    if err != nil {
      fmt.Println("Error parsing JSON:", err)
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

    //printSolTx(fromAddr, addr, tr.Result.Transaction.Message.AccountKeys[1], amountSent, sig)
    return amountSent, true
  }
  return 0, false
}

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

// fetchFullTransaction retrieves the complete Solana transaction from RPC
// (with proper Content-Type header + retry logic)
func fetchFullTransaction(signature string) interface{} {
	url := "https://api.mainnet-beta.solana.com"

	for attempt := 1; attempt <= 8; attempt++ {  // increased from 3
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
			fmt.Printf("⚠️ [RPC] Network error for %s (attempt %d)\n", signature[:12]+"...", attempt)
			time.Sleep(time.Duration(attempt*400) * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		var txResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&txResponse); err != nil {
			time.Sleep(time.Duration(attempt*300) * time.Millisecond)
			continue
		}

		if result, exists := txResponse["result"]; exists && result != nil {
			fmt.Printf("✅ [RPC] Full tx data received for %s\n", signature[:12]+"...")
			return result
		}

		fmt.Printf("⚠️ [RPC] Transaction %s not yet available (result=null) — attempt %d\n", signature[:12]+"...", attempt)
		time.Sleep(time.Duration(attempt*600) * time.Millisecond) // exponential backoff
	}

	fmt.Printf("❌ [RPC] Failed to fetch tx %s after 8 attempts\n", signature[:12]+"...")
	return nil
}
// End of fetchFUllTransaction

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
