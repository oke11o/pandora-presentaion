package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"mime"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var randGenerator *rand.Rand

func init() {
	randGenerator = rand.New(rand.NewSource(time.Now().UnixMicro()))
}

const (
	defaultPort = "8092"

	userCount          = 10
	userMultiplicator  = 1000
	orderMultiplicator = 100
)

var (
	keys          = map[string]int64{}
	keysMutex     sync.RWMutex
	auth200       = map[int64]uint64{}
	auth200Mutex  sync.Mutex
	auth400       atomic.Uint64
	auth500       atomic.Uint64
	list200       = map[int64]uint64{}
	list200Mutex  sync.Mutex
	list400       atomic.Uint64
	list500       atomic.Uint64
	order200      = map[int64]uint64{}
	order200Mutex sync.Mutex
	order400      atomic.Uint64
	order500      atomic.Uint64
)
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())

	keysMutex.Lock()
	for i := int64(1); i <= userCount; i++ {
		token := randStringRunes(64)
		keys[token+"-"+strconv.Itoa(int(i))] = i
	}
	keysMutex.Unlock()
}

type StatisticBodyResponse struct {
	Code200 map[int64]uint64 `json:"200"`
	Code400 uint64           `json:"400"`
	Code500 uint64           `json:"500"`
}

type StatisticResponse struct {
	Auth  StatisticBodyResponse `json:"auth"`
	List  StatisticBodyResponse `json:"list"`
	Order StatisticBodyResponse `json:"order"`
}

func checkContentTypeAndMethod(r *http.Request, methods []string) (int, error) {
	contentType := r.Header.Get("Content-Type")
	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return http.StatusBadRequest, errors.New("malformed Content-Type header")
	}

	if mt != "application/json" {
		return http.StatusUnsupportedMediaType, errors.New("header Content-Type must be application/json")
	}

	for _, method := range methods {
		if r.Method == method {
			return 0, nil
		}
	}
	return http.StatusMethodNotAllowed, errors.New("method not allowed")
}

func checkAuthorization(r *http.Request) (int64, int, error) {
	authHeader := r.Header.Get("Authorization")
	authHeader = strings.Replace(authHeader, "Bearer ", "", 1)
	keysMutex.RLock()
	userID := keys[authHeader]
	keysMutex.RUnlock()

	if userID == 0 {
		return 0, http.StatusUnauthorized, errors.New("StatusUnauthorized")
	}
	return userID, 0, nil
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	code, err := checkContentTypeAndMethod(r, []string{http.MethodPost})
	if err != nil {
		if code >= 500 {
			auth500.Add(1)
		} else {
			auth400.Add(1)
		}
		http.Error(w, err.Error(), code)
		return
	}

	user := struct {
		UserID int64 `json:"user_id"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		auth500.Add(1)
		http.Error(w, "Incorrect body", http.StatusNotAcceptable)
		return
	}
	if user.UserID > userCount {
		auth400.Add(1)
		http.Error(w, "Incorrect user_id", http.StatusBadRequest)
		return
	}

	auth200Mutex.Lock()
	auth200[user.UserID]++
	auth200Mutex.Unlock()

	var authKey string
	keysMutex.RLock()
	for k, v := range keys {
		if v == user.UserID {
			authKey = k
			break
		}
	}
	keysMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(fmt.Sprintf(`{"auth_key": "%s"}`, authKey)))
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	code, err := checkContentTypeAndMethod(r, []string{http.MethodGet})
	if err != nil {
		if code >= 500 {
			list500.Add(1)
		} else {
			list400.Add(1)
		}
		http.Error(w, err.Error(), code)
		return
	}

	userID, code, err := checkAuthorization(r)
	if err != nil {
		list400.Add(1)
		http.Error(w, err.Error(), code)
		return
	}

	list200Mutex.Lock()
	list200[userID]++
	list200Mutex.Unlock()

	// Logic
	userID *= userMultiplicator
	result := make([]string, orderMultiplicator)
	for i := int64(0); i < orderMultiplicator; i++ {
		result[i] = strconv.FormatInt(userID+i, 10)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(fmt.Sprintf(`{"items": [%s]}`, strings.Join(result, ","))))
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	code, err := checkContentTypeAndMethod(r, []string{http.MethodPost})
	if err != nil {
		if code >= 500 {
			order500.Add(1)
		} else {
			order400.Add(1)
		}
		http.Error(w, err.Error(), code)
		return
	}

	userID, code, err := checkAuthorization(r)
	if err != nil {
		list400.Add(1)
		http.Error(w, err.Error(), code)
		return
	}

	// Logic
	itm := struct {
		ItemID int64 `json:"item_id"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&itm)
	if err != nil {
		order500.Add(1)
		http.Error(w, "Incorrect body", http.StatusNotAcceptable)
		return
	}

	ranger := userID * userMultiplicator
	if itm.ItemID < ranger || itm.ItemID >= ranger+orderMultiplicator {
		order400.Add(1)
		http.Error(w, "Incorrect user_id", http.StatusBadRequest)
		return
	}

	order200Mutex.Lock()
	order200[userID]++
	order200Mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(fmt.Sprintf(`{"order": %d}`, itm.ItemID+12345)))
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	auth200Mutex.Lock()
	auth200 = map[int64]uint64{}
	auth200Mutex.Unlock()
	auth400.Store(0)
	auth500.Store(0)

	list200Mutex.Lock()
	list200 = map[int64]uint64{}
	list200Mutex.Unlock()
	list400.Store(0)
	list500.Store(0)

	order200Mutex.Lock()
	order200 = map[int64]uint64{}
	order200Mutex.Unlock()
	order400.Store(0)
	order500.Store(0)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

func statisticHandler(w http.ResponseWriter, r *http.Request) {
	response := StatisticResponse{
		Auth: StatisticBodyResponse{
			Code200: auth200,
			Code400: auth400.Load(),
			Code500: auth500.Load(),
		},
		List: StatisticBodyResponse{
			Code200: list200,
			Code400: list400.Load(),
			Code500: list500.Load(),
		},
		Order: StatisticBodyResponse{
			Code200: order200,
			Code400: order400.Load(),
			Code500: order500.Load(),
		},
	}
	b, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

func SleepMW(originalHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, vv := range r.URL.Query() {
			if k == "sleep" && len(vv) == 1 {
				if sleep, err := strconv.Atoi(vv[0]); err == nil && sleep > 0 && sleep <= 2000 {
					time.Sleep(time.Duration(sleep) * time.Millisecond)
				}
			}
		}
		originalHandler.ServeHTTP(w, r)
	})
}

/*
/url?fail=1000  - means that each 1000 requests will failed
*/
func FailImitationMW(originalHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, vv := range r.URL.Query() {
			if k != "fail" {
				continue
			}
			if len(vv) != 1 {
				break
			}
			chance, err := strconv.Atoi(vv[0])
			if err != nil || chance <= 0 {
				break
			}
			if randGenerator.Intn(chance) != 0 {
				break
			}

			userID, code, err := checkAuthorization(r)
			if err != nil {
				list400.Add(1)
				http.Error(w, err.Error(), code)
				return
			}
			_ = userID

			rnd := randGenerator.Intn(3)
			switch rnd {
			case 0:
				m := make(map[string]any)
				b, err := json.Marshal(m)
				if err != nil {
					list500.Add(1)
					http.Error(w, "Internal error", http.StatusInternalServerError)
					return
				}
				list200Mutex.Lock()
				list200[userID]++
				list200Mutex.Unlock()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(b)
			case 1:
				list400.Add(1)
				http.Error(w, "Bad request", http.StatusBadRequest)
			case 2:
				list500.Add(1)
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
			return
		}
		originalHandler.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	mux := http.NewServeMux()

	mux.Handle("/auth", SleepMW(FailImitationMW(http.HandlerFunc(authHandler))))
	mux.Handle("/list", SleepMW(FailImitationMW(http.HandlerFunc(listHandler))))
	mux.Handle("/order", SleepMW(FailImitationMW(http.HandlerFunc(orderHandler))))
	mux.Handle("/statistic", http.HandlerFunc(statisticHandler))
	mux.Handle("/reset", http.HandlerFunc(resetHandler))

	ctx := context.Background()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	log.Printf("Auth keys: %v", keys)
	log.Printf("Listening on :%s...", port)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}
	log.Fatal(err)
}
