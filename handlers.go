package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type Handlers struct {
	maxDuplicateCallbacks   int
	maxOrderProcessDuration int
	orderProcessRepo        *OrderProcessRepository
}

type HandlersCfg struct {
	MaxDuplicateCallbacks   int
	MaxOrderProcessDuration int
	OrderProcessRepo        *OrderProcessRepository
}

func NewHandlers(cfg HandlersCfg) *Handlers {
	return &Handlers{
		maxDuplicateCallbacks:   cfg.MaxDuplicateCallbacks,
		maxOrderProcessDuration: cfg.MaxOrderProcessDuration,
		orderProcessRepo:        cfg.OrderProcessRepo,
	}
}

func (h *Handlers) GetOrderProcess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		orderProcesses, err := h.orderProcessRepo.GetAll()
		if err != nil {
			log.Printf("error getting order processes: %#v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error"}`))
			return
		}

		if err = json.NewEncoder(w).Encode(orderProcesses); err != nil {
			log.Printf("error encoding order processes json: %#v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error"}`))
			return
		}

		return
	}

	orderProcess, err := h.orderProcessRepo.GetByOrderID(orderID)
	if err != nil {
		if err == ErrOrderProcessDoesNotExist {
			log.Printf("order process for order %s does not exist", orderID)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
			return
		}

		log.Printf("error getting order process by order id: %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error"}`))
		return
	}

	if err = json.NewEncoder(w).Encode(orderProcess); err != nil {
		log.Printf("error encoding order process json: %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error"}`))
		return
	}
}

func (h *Handlers) StartOrderProcess(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"invalid content type. Valid: application/json"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var orderProcess OrderProcess
	if err := json.NewDecoder(r.Body).Decode(&orderProcess); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"invalid request body"}`))
		return
	}
	defer r.Body.Close()

	if orderProcess.OrderID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"missing order id"}`))
		return
	}

	orderProcess.Status = STATUS_RUNNING
	if err := h.orderProcessRepo.Add(orderProcess); err != nil {
		if err == ErrOrderProcessExists {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"order process for order ` + orderProcess.OrderID + ` exists"}`))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error"}`))
		return
	}

	callbackURLStr := r.URL.Query().Get("callback_url")
	if callbackURLStr == "" {
		log.Printf("processing order %s with no callback\n", orderProcess.OrderID)
		go h.simulateOrderProcessing(orderProcess, "")
		return
	}

	callbackURL, err := url.ParseRequestURI(callbackURLStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"invalid callback_url: "` + callbackURLStr + `}`))
		return
	}

	log.Printf("processing order %s with callback\n", orderProcess.OrderID)
	go h.simulateOrderProcessing(orderProcess, callbackURL.String())

	w.WriteHeader(http.StatusAccepted)
}

// simulate a long running process and randomly assign a status to an order process
func (h *Handlers) simulateOrderProcessing(orderProcess OrderProcess, callbackURL string) {
	orderProcessFailed := rand.Intn(2) == 1
	orderProcessDuration := time.Duration(rand.Intn(h.maxOrderProcessDuration))
	orderProcessCallbackDuplications := rand.Intn(h.maxDuplicateCallbacks)

	log.Printf("order process for order %s will take %d seconds\n", orderProcess.OrderID, orderProcessDuration)
	time.Sleep(orderProcessDuration * time.Second)

	if orderProcessFailed {
		orderProcess.Status = STATUS_FAILED
	} else {
		orderProcess.Status = STATUS_SUCCEEDED
	}
	log.Printf("order process for order %s has status %s\n", orderProcess.OrderID, orderProcess.Status)

	if err := h.orderProcessRepo.Update(orderProcess); err != nil {
		log.Fatalf("error updating order process: %q", err)
	}

	if callbackURL == "" {
		return // empty string indicates that no callback should be made...
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(orderProcess); err != nil {
		log.Fatalf("error encoding order process json: %q", err)
	}

	log.Printf("callback for order %s will be duplicated %d times\n", orderProcess.OrderID, orderProcessCallbackDuplications)

	for i := 0; i <= orderProcessCallbackDuplications; i++ {
		var retries int
		for {
			if retries == 3 {
				log.Printf("unable to callback for order: %q. Abandoning callback.\n", orderProcess.OrderID)
				return
			}

			resp, err := http.Post(callbackURL, "application/json", &buf)
			if err != nil {
				log.Printf("error calling callback url: %q. Retrying in 5 seconds...\n", err)
				time.Sleep(5 * time.Second)
				retries++
				continue
			}
			if resp.StatusCode != http.StatusOK {
				log.Printf("error calling callback url: Received status code %d. Retrying in 5 seconds...\n", resp.StatusCode)
				time.Sleep(5 * time.Second)
				retries++
				continue
			}

			log.Printf("successfully called back for order %q duplicate %d\n", orderProcess.OrderID, i)
			break
		}
	}
}
