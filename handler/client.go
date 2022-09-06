package handler

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/sys/cpu"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"websocket-splice/models"
)

var (
	platform       = runtime.GOOS
	filename       = "scanner" + time.Now().Format("-2006-01-02-15-04-05T-07-00")
	timeout        int
	timeoutService int
	creditsNow     int
)

func Start() error {
	switch platform {
	case "windows":
		platform = "Windows"
	case "linux":
		platform = "Linux"
	case "darwin":
		platform = "Macintosh"
	}

	if platform == "" {
		return errors.New("can not start. invalid os info")
	}
	C = make(map[int]*websocket.Conn)
	newPath = filepath.Join(os.TempDir(), "Splice Scanner")

	LogStart()
	LogCommon(fmt.Sprintf(`platform %s`, platform))

	arch := runtime.GOARCH
	LogCommon(fmt.Sprintf(`arch %s`, arch))
	LogCommon(fmt.Sprintf(`cores %d`, runtime.NumCPU()))
	if arch == "amd64" || arch == "386" {
		LogCommon(fmt.Sprintf(`sse2 %v`, cpu.X86.HasSSE2))
		LogCommon(fmt.Sprintf(`sse3 %v`, cpu.X86.HasSSE3))
		LogCommon(fmt.Sprintf(`sse41 %v`, cpu.X86.HasSSE41))
		LogCommon(fmt.Sprintf(`sse42 %v`, cpu.X86.HasSSE42))
		LogCommon(fmt.Sprintf(`avx %v`, cpu.X86.HasAVX))
		LogCommon(fmt.Sprintf(`avx2 %v`, cpu.X86.HasAVX2))
	}
	if arch == "arm64" {
		LogCommon(fmt.Sprintf(`asimddp %v`, cpu.ARM64.HasASIMDDP))
		LogCommon(fmt.Sprintf(`asimdfhm %v`, cpu.ARM64.HasASIMDFHM))
		LogCommon(fmt.Sprintf(`asimd %v`, cpu.ARM64.HasASIMD))
		LogCommon(fmt.Sprintf(`asimdrdm %v`, cpu.ARM64.HasASIMDRDM))
		LogCommon(fmt.Sprintf(`atomics %v`, cpu.ARM64.HasATOMICS))
	}

	return nil
}

func startCommand(client *http.Client, id int, session, agent string) {
	if timeout > 2 || timeoutService > 2 {
		messageNoResponse := "server not response. shutdown..."
		LogError(messageNoResponse)
		SendMessage(messageNoResponse)
		state = false
		timeout = 0
		timeoutService = 0
		return
	}

	log.Printf("start %d", id)

	var last models.InputLast

	lastBody, err := ExtractBody(client, "GET", host+"/api/promo/last", session, nil)
	if err != nil {
		timeout++
		LogError(fmt.Sprintf(`lastBody %v`, err))
		return
	}
	timeout = 0
	if err := json.Unmarshal(lastBody, &last); err != nil {
		LogError(err.Error())
		return
	}

	// Roll promo next if not pending
	if !last.Pending {
		next, err := promoStep(last.Promo, true)
		if err != nil {
			LogError(err.Error())
		}
		last.Promo = next
	}
	messageLast := fmt.Sprintf("Get Last: %s", shorter(last.Hash))
	LogCommon(messageLast)
	SendMessage(messageLast)

	pending := models.OutputPromo{
		Pending: last.Pending,
		Promo:   last.Promo,
		Hash:    last.Hash,
	}
	outPending, err := json.Marshal(pending)
	if err != nil {
		LogError(err.Error())
		return
	}

	_, err = ExtractBody(client, "POST", host+"/api/promo/pending", session, bytes.NewBuffer(outPending))
	if err != nil {
		timeout++
		LogError(fmt.Sprintf(`pendingBody %v`, err))
		return
	}
	timeout = 0
	SendMessage("Start Pending")

	var output models.OutputService

	for {
		serviceResponse, statusCode, err := ExtractServiceBody(client, "GET", service+"/www/payments/promo_codes/"+last.Promo, cookie.Cookie, agent)
		if err != nil {
			timeoutService++
			LogError(err.Error())
			return
		}
		timeoutService = 0
		strServiceResponse := string(serviceResponse)

		output = models.OutputService{
			Promo:  last.Promo,
			Info:   strServiceResponse,
			Status: "unexpected",
		}

		if statusCode < 300 {
			output.Status = "active"
		} else {
			if strings.Contains(strServiceResponse, "expired") {
				output.Status = "expired"
				break
			}
			if strings.Contains(strServiceResponse, "not found") {
				output.Status = "disabled"
				break
			}
			if strings.Contains(strServiceResponse, "Bad request") {
				cookie, err = GetCookie(client, login, password, agent)
				if err != nil {
					LogError(fmt.Sprintf("GetCookie %v", err))
					return
				}
				fmt.Println(cookie.Cookie)
				continue
			} else {
				break
			}
		}
	}

	outUpdate, err := json.Marshal(output)
	if err != nil {
		LogError(err.Error())
		return
	}

	encode, err := encryptAES(outUpdate, Token32)
	if err != nil {
		LogError(err.Error())
		return
	}

	outEncodeUpdate, err := json.Marshal(models.OutputUpdate{Hash: encode})
	if err != nil {
		LogError(err.Error())
		return
	}

	SendMessage("Updating...")

	updateResponse, err := ExtractBody(client, "PUT", host+"/api/promo/update", session, bytes.NewBuffer(outEncodeUpdate))
	if err != nil {
		timeout++
		LogError(fmt.Sprintf(`updateResponse %v`, err))
		return
	}
	timeout = 0

	var update models.InputAccount
	if err := json.Unmarshal(updateResponse, &update); err != nil {
		LogError(fmt.Sprintf(`Update Failed. %v`, err))
		return
	}

	if update.Credits != creditsNow {
		messageUpdate := fmt.Sprintf("Update Success. Credits: %d", update.Credits)
		LogCommon(messageUpdate)
		SendMessage(messageUpdate)
		//time.Sleep(500 * time.Millisecond)
		creditsNow = update.Credits
	}
}

func shorter(i string) string {
	if len(i) < 8 {
		return i
	}

	return fmt.Sprintf("%s...%s", i[:4], i[len(i)-4:])
}

func encryptAES(text []byte, key []byte) (string, error) {
	// generate a new aes cipher using our 32 byte long key
	c, err := aes.NewCipher(key)
	// if there are any errors, handle them
	if err != nil {
		return "", err
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	// if any error generating new GCM
	// handle them
	if err != nil {
		return "", err
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// here we encrypt our text using the Seal function
	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	output := gcm.Seal(nonce, nonce, text, nil)

	// return hexadecimal string
	return base64.StdEncoding.EncodeToString(output), nil
}
