package module

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/labstack/echo/v5"
	"golang.org/x/exp/constraints"
)

func IsNilOrEmptyString(value *string) bool {
	if value == nil {
		return true
	}

	return IsEmptyString(*value)
}

func IsEmptyString(value string) bool {
	return len(CleanString(value)) == 0
}

func CleanString(value string) string {
	return html.EscapeString(strings.TrimSpace(value))
}

func CleanInternalString(value string) string {
	words := strings.Fields(value)

	return strings.Join(words, " ")
}

func CleanFilterString(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(value, ":", "_"), "{", "("), "}", ")")
}

func DataString(data interface{}) string {
	out, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(out)
}

func StrToIntDefault(val string, def int) int {
	newVal, err := strconv.Atoi(val)
	if err != nil {
		newVal = def
	}

	return newVal
}

func GetParamDefault(val string, def string) string {
	if IsEmptyString(val) {
		val = def
	}

	return val
}

func RemovePrefixMobile(areaCode int, mobile string) string {
	if ((areaCode == 0) || (areaCode == 84)) && (strings.HasPrefix(mobile, "0")) {
		mobile = mobile[1:]
	}

	return mobile
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func IsArray(v interface{}) bool {
	if v == nil {
		return false
	}

	value := reflect.ValueOf(v)
	kind := value.Kind()

	return kind == reflect.Slice || kind == reflect.Array
}

func GetClientIP(r *http.Request) string {
	if ip := r.Header.Get(echo.HeaderXForwardedFor); !IsEmptyString(ip) {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}

	if ip := r.Header.Get(echo.HeaderXRealIP); !IsEmptyString(ip) {
		return ip
	}

	if ip := r.Header.Get("CF-Connecting-IP"); !IsEmptyString(ip) {
		return ip
	}

	return strings.TrimSpace(r.RemoteAddr)
}

func RoundUp(value float64, decimalPlaces int) float64 {
	factor := math.Pow10(decimalPlaces)
	scaled := value * factor
	ceil := math.Ceil(scaled)

	return ceil / factor
}

func Round(value float64, decimalPlaces int) float64 {
	factor := math.Pow10(decimalPlaces)
	scaled := value * factor
	rounded := math.Round(scaled)

	return rounded / factor
}

func RoundDown(value float64, decimalPlaces int) float64 {
	factor := math.Pow10(decimalPlaces)
	scaled := value * factor
	floor := math.Floor(scaled)

	return floor / factor
}

// SliceToMap converts a slice of any type to a map.
// The key for the map is extracted from each element using the provided keyFunc.
// The key must be comparable.
func SliceToMap[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	if len(slice) == 0 {
		return make(map[K]T)
	}

	mapResult := make(map[K]T)
	for _, item := range slice {
		key := keyFunc(item)
		mapResult[key] = item
	}

	return mapResult
}

func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))

	for k, v := range a {
		out[k] = v
	}

	for k, v := range b {
		// If you use map[string]interface{}, ok is always false here.
		// Because yaml.Unmarshal will give you map[interface{}]interface{}.
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}

	return out
}

func FuncName() string {
	// Skip this function, and fetch the PC and file for its parent.
	pc, _, _, _ := runtime.Caller(1)

	// Retrieve a function object this functions parent.
	funcObj := runtime.FuncForPC(pc)

	// Regex to extract just the function name (and not the module path).
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)

	return runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")
}

// The Ordered constraint is defined in golang.org/x/exp/constraints
// and includes all basic types that support the <, >, <=, >= operators:
// Ordered = Integer | Float | ~string
type Ordered = constraints.Ordered

// Define the generic function with the new Ordered constraint.
func GetMinNItems[T any, C Ordered](items []T, n int, accessor func(T) C) []T {
	// Check if the input is empty or N is zero
	if len(items) == 0 || n <= 0 {
		return nil
	}

	// Create a copy of the slice to avoid modifying the original
	itemsCopy := make([]T, len(items))
	copy(itemsCopy, items)

	// Custom sort the slice based on the value returned by the accessor function
	// The comparison requires the type C to support the '<' operator.
	// This relies on Go's internal handling of comparable types.
	// Note: For non-numeric comparable types like strings, '<' performs lexicographical comparison.
	sort.Slice(itemsCopy, func(i, j int) bool {
		valI := accessor(itemsCopy[i])
		valJ := accessor(itemsCopy[j])
		return valI < valJ // Ascending sort (minimum first)
	})

	// Return the first N elements
	if len(itemsCopy) < n {
		return itemsCopy
	}

	return itemsCopy[:n]
}

func GetUrl(domain string, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	path = strings.TrimPrefix(path, "/")

	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return fmt.Sprintf("%s/%s", domain, path)
	}

	return fmt.Sprintf("https://%s/%s", domain, path)
}

func MappingData(datas map[string]interface{}, key string, data any) map[string]interface{} {
	if datas == nil {
		datas = make(map[string]interface{})
	}

	datas[key] = data

	return datas
}

func MaskingString(input string, firstVisible int, lastVisible int) string {
	inputLen := len(input)
	if inputLen == 0 {
		return ""
	}

	const minMaskLen = 2

	// Pattern: Only Last X Visible (firstVisible = 0)
	if firstVisible <= 0 {
		mLen := max(minMaskLen, inputLen-lastVisible)
		mLen = min(mLen, inputLen)
		return strings.Repeat("*", mLen) + input[mLen:]
	}

	// Pattern: Middle Masking (First and Last visible)
	f, l := firstVisible, lastVisible

	if inputLen < (f + l + minMaskLen) {
		allowedEdges := max(0, inputLen-minMaskLen)
		totalRequested := float64(f + l)

		if totalRequested > 0 {
			f = int(float64(f) / totalRequested * float64(allowedEdges))
			l = allowedEdges - f
		} else {
			f, l = 0, 0
		}
	}

	middleLen := inputLen - f - l
	return input[:f] + strings.Repeat("*", middleLen) + input[inputLen-l:]
}

func IsValidCryptoAddress(address string) bool {
	// 1. Check EVM (Ethereum/BSC/Polygon)
	// Pattern: 0x + 40 hex chars
	if regexp.MustCompile("^0x[0-9a-fA-F]{40}$").MatchString(address) {
		return true
	}

	// 2. Check Base58Check (Legacy BTC, Doge, Tron)
	// This decodes AND verifies the 4-byte double-SHA256 checksum
	_, _, err := base58.CheckDecode(address)
	if err == nil {
		return true
	}

	// 3. Check Bech32/Bech32m (SegWit BTC, Cosmos, Polkadot)
	_, _, err = bech32.Decode(address)
	if err == nil {
		return true
	}

	// 4. Check Solana (Base58 Raw)
	// Solana doesn't use a checksum, only a length check (32-44 chars)
	if regexp.MustCompile("^[1-9A-HJ-NP-Za-km-z]{32,44}$").MatchString(address) {
		return true
	}

	return false
}

func IsHaveBadWords(value string) bool {
	badWords := []string{
		"人大",
		"遊行",
		"六四",
		"天安門",
		"中共",
		"中華民國",
		"ROC",
		"China",
		"馬英九",
		"賴清德",
		"柯文哲",
		"蔡英文",
		"習近平",
		"鄧小平",
		"陳水扁",
		"吳敦義",
		"一國兩制",
		"一邊一國",
		"台獨",
		"台灣獨立",
		"共產黨",
		"國民黨",
		"民進黨",
		"主席",
		"政治",
		"屠殺",
		"子彈",
		"學運",
		"革命",
		"連戰",
		"宋楚瑜",
		"連勝文",
		"行政院",
		"蘇貞昌",
		"陳菊",
		"文革",
		"紅衛兵",
		"林彪",
		"江青",
		"劉少奇",
		"希特勒",
		"墨索里尼",
		"東條英機",
		"天皇",
		"共匪",
		"大日本帝國",
		"納粹",
		"鎮壓",
		"Obama",
		"倭寇",
		"日本鬼子",
		"周恩來",
		"馬克思",
		"恩格斯",
		"列寧",
		"史達林",
		"布希",
		"支那",
		"集中營",
		"毛澤東",
		"Nigger",
		"黑鬼",
		"基佬",
		"幹",
		"e04",
		"SHIT",
		"crap",
		"bullshit",
		"操",
		"艸",
		"肏",
		"糙",
		"靠",
		"凸",
		"機掰",
		"雞掰",
		"G8",
		"78",
		"雞歪",
		"雞雞歪歪",
		"機歪",
		"機機歪歪",
		"機車",
		"破麻",
		"你媽",
		"他媽",
		"你他媽",
		"你娘",
		"老母",
		"老木",
		"林涼",
		"拎涼",
		"暗陰陽",
		"按音羊",
		"mother fucker",
		"你妹",
		"媽的",
		"屎",
		"大便",
		"SHIT",
		"傻B",
		"傻逼",
		"嫩B",
		"嫩逼",
		"ON9",
		"TNND",
		"NTMD",
		"TMD",
		"MDFK",
		"SB",
		"日你",
		"神經病",
		"王八蛋",
		"asshole",
		"白痴",
		"低能",
		"弱智",
		"智障",
		"智缺",
		"moron",
		"jerk",
		"混蛋",
		"雜碎",
		"雜種",
		"狗雜種",
		"狗娘養的",
		"去死",
		"癡漢",
		"婊",
		"婊子",
		"女表",
		"cunt",
		"BITCH",
		"slut",
		"蕩婦",
		"賤",
		"貝戈戈",
		"Bastard",
		"幹你",
		"幹我",
		"幹他",
		"上你",
		"上我",
		"上他",
		"FUCK",
		"FK",
		"MAKE LOVE",
		"做愛",
		"愛愛",
		"打炮",
		"嘿咻",
		"性交",
		"sex",
		"coitus",
		"venery",
		"intercourse",
		"高潮",
		"orgasm",
		"勃起",
		"晨勃",
		"erection",
		"erect",
		"boner",
		"stiffy",
		"早洩",
		"石更",
		"射精",
		"身寸",
		"米青",
		"ejaculate",
		"ejaculation",
		"cum",
		"cumming",
		"精子",
		"精液",
		"中出",
		"內射",
		"creampie",
		"barebacking",
		"潤滑液",
		"潤滑油",
		"陰毛",
		"下體",
		"陽具",
		"屌",
		"陰莖",
		"肉棒",
		"老二",
		"雞雞",
		"陰囊",
		"睪丸",
		"蛋蛋",
		"懶趴",
		"懶叫",
		"龜頭",
		"馬眼",
		"penis",
		"dick",
		"cock",
		"ball",
		"潮吹",
		"squirt",
		"squirting",
		"squirter",
		"穴",
		"陰道",
		"pussy",
		"vagina",
		"vaginal",
		"陰部",
		"陰蒂",
		"clitoris",
		"clit",
		"A片",
		"成人片",
		"三級片",
		"色情",
		"情色",
		"AV",
		"PORN",
		"horny",
		"援交",
		"吃魚",
		"喝茶",
		"魚訊",
		"茶訊",
		"奶子",
		"咪咪",
		"內內",
		"奶奶",
		"女乃女乃",
		"Breast",
		"Busty",
		"激凸",
		"nipple",
		"乳房",
		"乳交",
		"性幻想",
		"裸體",
		"naked",
		"nude",
		"nudity",
		"3P",
		"threesome",
		"多P",
		"群P",
		"群交",
		"雜交",
		"多人",
		"騷",
		"騷逼",
		"SM",
		"bdsm",
		"性虐待",
		"內褲",
		"自慰",
		"masturbate",
		"masturbation",
		"打手槍",
		"jerk off",
		"jack off",
		"hand job",
		"wank",
		"18禁",
		"十八禁",
		"口交",
		"Blow job",
		"口爆",
		"口射",
		"口愛",
		"口交口交",
		"吞精",
		"獸交",
		"人獸",
		"處女",
		"virgin",
		"尻",
		"屁眼",
		"ass",
		"肛門",
		"肛交",
		"Anal",
		"AnalBeads",
		"Penetration",
		"淫",
		"意淫",
		"淫蕩",
		"發情",
		"媚藥",
		"淫亂",
		"色QQ",
		"色貓",
		"愛液",
		"猛插",
		"幹爆",
		"吸奶",
		"吹喇叭",
		"破處",
		"性高潮",
		"日死",
		"按摩棒",
		"dildo",
		"一本道",
		"東京熱",
		"tokyo hot",
		"指交",
		"fingering",
		"拳交",
		"analfisting",
		"顏射",
		"素人自拍",
		"Amateur",
		"吸毒",
		"drug",
		"毒品",
		"搖頭丸",
		"FM2",
		"安眠藥",
		"嗑藥",
		"販毒",
		"姦",
		"奸",
		"強姦",
		"強奸",
		"強暴",
		"rape",
		"rapist",
		"殺人",
		"kill",
		"homicide",
		"murder",
		"貪污",
		"corruption",
		"殺手",
		"killer",
		"偷拍",
		"作弊",
		"cheat",
		"老虎機",
		"釣魚機",
		"賭博",
		"gamble",
		"gambling",
		"wager",
		"運彩",
		"簽賭",
		"下注",
		"澳門首家",
		"自殺",
		"suicide",
		"大麻",
		"marijuana",
		"weed",
		"cannabis",
		"hemp",
		"pot",
		"罌粟",
		"poppy",
		"Papaver somniferum",
		"安非他命",
		"methamphetamine",
		"交易",
		"8591",
		"系統",
		"system",
		"GM",
		"Gamunity",
		"浩克娛樂",
		"hok entertainment",
		"官方",
		"official",
		"測試",
		"test",
		"QA",
		"http",
		"www",
		"com",
		"點數",
		"外掛",
		"CS",
		"客服",
		"公告",
		"announcement",
		"OP",
		"營運",
		"Operations",
	}

	// Create the regex pattern using the alternation operator '|'
	pattern := "(?i)("
	for i, s := range badWords {
		// Escape any special regex characters in the string
		pattern += regexp.QuoteMeta(s)
		if i < len(badWords)-1 {
			pattern += "|"
		}
	}
	pattern += ")"

	// Compile the regex
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return true
	}

	return regex.MatchString(value)
}
