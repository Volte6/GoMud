package util

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/md5"

	"github.com/volte6/mud/connection"
	"github.com/volte6/mud/term"
)

var (
	turnCount    uint64 = 0
	roundCount   uint64 = 0
	timeTrackers        = map[string]*Accumulator{}
	serverAddr   string = `Unknown`
	serverSeed   string = ``

	strippablePrepositions = []string{
		`onto`,
		`into`,
		`over`,
		`to`,
		`toward`,
		`towards`,
		`from`,
		`in`,
		`under`,
		`upon`,
		`with`,
		`the`, // also strip this because it's unnecessary
		`my`,  // also strip this because it's unnecessary
	}
)

const ()

type CommandQueue interface {
	QueueCommand(userId int, mobId int, cmd string, waitTurns ...int)
	QueueBuff(userId int, mobId int, buffId int)
	QueueQuest(userId int, questToken string)
	QueueRoomAction(roomId int, sourceUserId int, sourceMobId int, action string)
	Broadcast(msg string, skipLineRefresh ...bool)
	GetSettings(userId int) (connection.ClientSettings, error)
}

func SetServerAddress(addr string) {
	serverAddr = addr
}

func GetServerAddress() string {
	return serverAddr
}

func IncrementTurnCount() uint64 {
	turnCount++
	return turnCount
}

func GetTurnCount() uint64 {
	return turnCount
}

func IncrementRoundCount() uint64 {
	roundCount++
	return roundCount
}

func GetRoundCount() uint64 {
	return roundCount
}

func TrackTime(name string, timePassed float64) {
	if _, ok := timeTrackers[name]; !ok {
		timeTrackers[name] = &Accumulator{
			Name:  name,
			Start: time.Now()}
	}
	timeTrackers[name].Record(timePassed)
}

func GetTimeTrackers() []Accumulator {

	result := []Accumulator{}
	for _, t := range timeTrackers {
		result = append(result, *t)
	}

	return result
}

type Accumulator struct {
	Name    string
	Total   float64
	Lowest  float64
	Highest float64
	Count   float64
	Start   time.Time
}

func (t *Accumulator) Stats() (lowest float64, highest float64, average float64, count float64) {
	return t.Lowest, t.Highest, t.Average(), t.Count
}

func (t *Accumulator) Average() float64 {
	return t.Total / t.Count
}

func (t *Accumulator) Record(nextValue float64) {
	t.Count++
	t.Total += nextValue
	if nextValue < t.Lowest || t.Lowest == 0 {
		t.Lowest = nextValue
	}
	if nextValue > t.Highest {
		t.Highest = nextValue
	}
}

func Rand(maxInt int) int {
	if maxInt < 1 {
		return 0
	}
	return rand.Intn(maxInt)
}

func LogRoll(name string, rollResult int, targetNumber int) {
	success := rollResult < targetNumber
	slog.Info(`Rand Result`, `Name`, name, `Result`, fmt.Sprintf(`%d < %d`, rollResult, targetNumber), `Success`, success)
}

func SplitString(input string, lineWidth int) []string {
	var result []string
	words := strings.Fields(input) // Split the input into words

	currentLine := ""
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= lineWidth { // +1 for the space
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			result = append(result, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		result = append(result, currentLine)
	}

	return result
}

// Splits a string by adding line breaks at the end of each line
func SplitStringNL(input string, lineWidth int) string {

	output := strings.Builder{}

	words := strings.Fields(input) // Split the input into words

	currentLine := ""
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= lineWidth { // +1 for the space
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			output.WriteString(currentLine)
			output.WriteString(term.CRLFStr)
			currentLine = word
		}
	}

	if currentLine != "" {
		output.WriteString(currentLine)
	}

	return output.String()
}

func SplitButRespectQuotes(s string) []string {

	// This regex matches either a quoted string (with either single or double quotes) or a non-space sequence.
	// For example, for the input: `hello "my name" is 'Sammy'`
	// It matches: [`hello", ""my name"", "is", "'Sammy'`]
	re := regexp.MustCompile(`("[^"]*")|('[^']*')|\S+`)
	matches := re.FindAllString(s, -1)
	finalMatches := make([]string, 0, 1)

	// Remove quotes around the matches, if they exist
	for _, match := range matches {

		match = strings.TrimSpace(match)

		if strings.HasPrefix(match, `"`) && strings.HasSuffix(match, `"`) ||
			strings.HasPrefix(match, `'`) && strings.HasSuffix(match, `'`) {
			str := strings.TrimSpace(match[1 : len(match)-1])
			finalMatches = append(finalMatches, str)
		} else {
			finalMatches = append(finalMatches, match)
		}
	}

	return finalMatches
}

// accepts an input and splits it along a # if any.
// By default returns the full string and 1 as the number.
func GetMatchNumber(input string) (string, int) {
	// Clean up the input
	input = strings.TrimSpace(strings.ToLower(input))
	// See if the item has a # and if so grab the left as the name, and the right as the number
	if !strings.Contains(input, "#") {
		return input, 1
	}

	parts := strings.Split(input, "#")
	input = parts[0]
	inputNumber, _ := strconv.Atoi(strings.Join(parts[1:], "#"))
	if inputNumber < 1 {
		inputNumber = 1
	}

	return input, inputNumber
}

func FindMatchIn(searchName string, items ...string) (match string, closeMatch string) {

	if searchName == `` {
		return ``, `` // No match
	}

	searchName, searchNumber := GetMatchNumber(searchName)

	var matchCt int = 0
	var closeMatchCt int = 0

	for _, i := range items {

		part, full := stringMatch(searchName, i, false)

		if part {
			closeMatchCt++
			if closeMatchCt == searchNumber {
				closeMatch = i
			}
		}

		if full {
			matchCt++
			if matchCt == searchNumber {
				match = i
				break
			}
		}

	}

	// If no "starts with" or "exact" matches are found, try and find the first item that contain the supplied name
	// Note: Can't have an exact match if there was never a close match
	if len(closeMatch) == 0 {
		closeMatchCt = 0
		for _, i := range items {
			part, _ := stringMatch(searchName, i, true)

			if part {
				closeMatchCt++
				if closeMatchCt == searchNumber {
					closeMatch = i
					break
				}
			}

		}

	}

	return match, closeMatch
}

// Searches for a partial or full match of a string
// If allowContains is true, the match can appear anywhere in the string.
// Otherwise it must start with the searchFor string
func stringMatch(searchFor string, searchIn string, allowContains bool) (partialMatch bool, fullMatch bool) {

	searchFor = strings.ToLower(searchFor)
	searchIn = strings.ToLower(searchIn)

	if allowContains {
		if strings.Contains(searchIn, searchFor) {
			if searchIn == searchFor {
				return true, true
			}
			return true, false
		}
	}

	if strings.HasPrefix(searchIn, searchFor) {
		if searchIn == searchFor {
			return true, true
		}
		return true, false
	}

	return false, false
}

func Hash(input string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}

func HashBytes(input []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(input))
}

func Md5(input string) string {
	h := md5.New()
	return string(h.Sum([]byte(input)))
}

func Md5Bytes(input []byte) []byte {
	h := md5.New()
	return h.Sum(input)
}

func GetLockSequence(lockIdentifier string, difficulty int, seed string) string {
	// A lock sequence is a sequence of UP or DOWN commands that must be entered to unlock a lock
	// The difficulty is how many commands are in the sequence
	// First generate a Md5Bytes() hash and then use the first N bytes to generate the sequence
	// The sequence is generated by taking the first N bytes of the hash and then converting each byte to a number
	// If the number is even, the command is UP, if it's odd, the command is DOWN
	// The sequence is then returned as a string of U's and D's

	// Generate the hash
	hash := Md5Bytes([]byte(strings.ToLower(lockIdentifier + seed)))

	// Minimum difficulty of 2 (2^2 random chance)
	if difficulty < 2 {
		difficulty = 2
	}
	// Maxinum difficulty of 32 (2^32 random chance)
	if difficulty > 32 {
		difficulty = 32
	}

	// Generate the sequence
	sequence := ""

	for i := 0; i < difficulty; i++ {
		if hash[i]%2 == 0 {
			sequence += "U"
		} else {
			sequence += "D"
		}
	}

	return sequence
}

func Compress(input []byte) []byte {
	var b bytes.Buffer
	// Create a new gzip writer
	gz := gzip.NewWriter(&b)
	// Write the input data to the gzip writer
	if _, err := gz.Write(input); err != nil {
		return []byte{}
	}
	if err := gz.Close(); err != nil {
		return []byte{}
	}
	return b.Bytes()
}

func Decompress(input []byte) []byte {

	// Create a buffer to read from the compressed data
	b := bytes.NewBuffer(input)
	// Create a new gzip reader
	gr, err := gzip.NewReader(b)
	if err != nil {
		return []byte{}
	}
	defer gr.Close()

	// Read the uncompressed data from the gzip reader
	uncompressedData, err := io.ReadAll(gr)
	if err != nil {
		return []byte{}
	}

	return uncompressedData
}

func Encode(blobdata []byte) string {
	// base64 encode the bytes
	return base64.StdEncoding.EncodeToString(blobdata)
}

func Decode(base64str string) []byte {
	// base64 encode the bytes
	b, _ := base64.StdEncoding.DecodeString(base64str)
	return b
}

func GetMyIP() string {

	url := `https://api.ipify.org/?format=txt`

	resp, err := http.Get(url)
	if err != nil {
		return err.Error()
	}

	defer resp.Body.Close()
	// handle the error if there is one
	if err != nil {
		return err.Error()
	}

	// do this now so it won't be forgotten
	defer resp.Body.Close()
	// reads html as a slice of bytes
	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}

	// show the HTML code as a string %s
	return string(html)
}

func ProgressBar(complete float64, maxBarSize int, barParts ...string) (fullBar string, emptyBar string) {
	fullBarPiece := `█`
	emptyBarPiece := `░`

	if len(barParts) >= 2 {
		fullBarPiece = barParts[0]
		emptyBarPiece = barParts[1]
	}

	fullBars := int(math.Floor(float64(maxBarSize) * complete))
	return strings.Repeat(fullBarPiece, fullBars), strings.Repeat(emptyBarPiece, maxBarSize-fullBars)
}

// Returns X dice rolled with Y sides
func RollDice(dice int, sides int) int {
	var total int

	invert := dice < 0

	if invert {
		dice *= -1
	}

	if sides < 0 {
		sides *= -1
	}

	for i := 0; i < dice; i++ {
		total += Rand(sides) + 1
	}

	if invert {
		return total * -1
	}

	return total
}

// Gets the specifics of the item damage
// Format:
// 2@1d3+2#1,2,3
func ParseDiceRoll(dRoll string) (attacks int, dCount int, dSides int, bonus int, buffOnCrit []int) {

	attacks = 1

	var dice []string

	// After # is a list of buffId's separated by commas
	if strings.Contains(dRoll, `#`) {
		parts := strings.Split(dRoll, `#`)
		dRoll = parts[0]

		buffIds := strings.Split(parts[1], `,`)
		for _, buffId := range buffIds {
			buffId = strings.TrimSpace(buffId)
			buffIdInt, _ := strconv.Atoi(buffId)
			if buffIdInt != 0 {
				buffOnCrit = append(buffOnCrit, buffIdInt)
			}
		}
	}

	invertCount := 1
	if dRoll[0] == '-' {
		dRoll = strings.TrimLeft(dRoll, `-`)
		invertCount = -1
	}

	// 1d3+2, 1d3-1, etc
	// Determine if the bonus is negative or positive
	bonusFactor := 1
	if strings.Contains(dRoll, `+`) {
		dice = strings.Split(dRoll, `+`)
	} else if strings.Contains(dRoll, `-`) {
		bonusFactor = -1 // Invert the bonus
		dice = strings.Split(dRoll, `-`)
	} else {
		dice = []string{dRoll}
	}

	// Apply bonus
	if len(dice) == 2 {
		dice[1] = strings.TrimSpace(dice[1])
		bonus, _ = strconv.Atoi(dice[1])
		bonus *= bonusFactor
	}

	// Parse the dice details
	die := dice[0]

	// How many times does this dice roll get?
	// Only override attacks if we have a valid attack argument provided
	// 2@1d3+2 etc
	attackParts := strings.Split(die, `@`)
	if len(attackParts) == 2 {
		attacks, _ = strconv.Atoi(attackParts[0])
		die = attackParts[1]
	}

	// 2d4 etc.
	dieParts := strings.Split(die, `d`)
	if len(dieParts) == 2 {

		dieParts[0] = strings.TrimSpace(dieParts[0])
		dieParts[1] = strings.TrimSpace(dieParts[1])

		dCount, _ = strconv.Atoi(dieParts[0])
		dSides, _ = strconv.Atoi(dieParts[1])
	}

	return attacks, invertCount * dCount, dSides, bonus, buffOnCrit
}

func FormatDiceRoll(attacks int, dCount int, dSides int, bonus int, buffOnCrit []int) string {

	dRoll := ``

	// 2@
	if attacks != 1 {
		dRoll = fmt.Sprintf(`%d@`, attacks)
	}

	// 2d6
	dRoll += fmt.Sprintf(`%dd%d`, dCount, dSides)

	// +2
	if bonus != 0 {
		if bonus > 0 {
			dRoll += fmt.Sprintf(`+%d`, bonus)
		} else {
			dRoll += fmt.Sprintf(`-%d`, bonus*-1)
		}
	}

	// #9,11,30
	if len(buffOnCrit) > 0 {
		for _, buffId := range buffOnCrit {
			dRoll = fmt.Sprintf(`%s#%d,`, dRoll, buffId)
		}
		dRoll = strings.TrimRight(dRoll, `,`)
	}

	return dRoll
}

// SafeSave first saves to a temp file, then renames it to save over the target destination
// This is to lessen the risk of a partial write being interrupted and corrupting the file
// due to power loss etc.
func SafeSave(path string, data []byte) error {

	path = filepath.FromSlash(path)

	safePath := path + `.new`

	if err := os.WriteFile(safePath, data, 0777); err != nil {
		return err
	}

	//
	// Once the file is written, rename it to remove the .new suffix and overwrite the old file
	//
	if err := os.Rename(safePath, path); err != nil {
		return err
	}

	return nil
}

// Basic save wrapper
func Save(path string, data []byte, doSafe ...bool) error {

	path = filepath.FromSlash(path)

	if len(doSafe) > 0 && doSafe[0] {
		return SafeSave(path, data)
	}

	if err := os.WriteFile(path, data, 0777); err != nil {
		return err
	}

	return nil
}

func FilePath(pathParts ...string) string {
	if len(pathParts) == 1 {
		return filepath.FromSlash(pathParts[0])
	}
	return filepath.FromSlash(strings.Join(pathParts, ``))
}

func BreakIntoParts(full string) []string {
	result := []string{full}

	parts := strings.Split(full, ` `)
	partCt := len(parts)
	for i := 1; i < partCt; i++ {
		result = append(result, strings.Join(parts[i:], ` `))
	}

	return result
}

func HealthClass(health int, maxHealth int) string {

	if health <= 0 {
		return `health-dead`
	}
	// quantize to 10s
	healthPercent := int(math.Floor(float64(health)/float64(maxHealth)*10)) * 10

	return fmt.Sprintf(`health-%d`, healthPercent)
}

func ManaClass(mana int, maxMana int) string {

	// quantize to 10s
	manaPercent := int(math.Floor(float64(mana)/float64(maxMana)*10)) * 10

	return fmt.Sprintf(`mana-%d`, manaPercent)
}

func QuantizeTens(value int, max int) int {
	return int(math.Floor(float64(value)/float64(max)*10)) * 10
}

// Strips out common prepositions from a string
func StripPrepositions(input string) string {

	if input == `` {
		return input
	}

	for _, prep := range strippablePrepositions {
		prepLen := len(prep)

		if len(input) > prepLen && input[0:len(prep)+1] == prep+` ` {
			input = input[len(prep)+1:]
		}
		input = strings.ReplaceAll(input, ` `+prep+` `, ` `)
	}

	return input
}
