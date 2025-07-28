package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Row struct {
	JobID string `json:"job_id"`
	Data string `json:"data"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func streamHandler(w http.ResponseWriter, r *http.Request) {

	rows := make([]Row, 0)
	for i := 0; i < 1000000; i++ {
		rows = append(rows, Row{
			JobID: fmt.Sprintf("job_%d", i),
			Data: fmt.Sprintf("data_%d", i),
			Status: "pending",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	boundary := "boundary123abc"
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(200)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	sendPart := func(jsonPayload string) {
		fmt.Fprintf(w, "--%s\r\n", boundary)
		fmt.Fprint(w, "Content-Type: application/json\r\n\r\n")
		fmt.Fprint(w, jsonPayload)
		fmt.Fprint(w, "\r\n")
		flusher.Flush()
	}
	
	for _, row := range rows {
		json, err := json.Marshal(row)
		if err != nil {
			sendPart(fmt.Sprintf(`{"job_id": "%s", "status": "error", "error": "%s"}`, row.JobID, err.Error()))
			return
		}
		sendPart(string(json))
		time.Sleep(500 * time.Millisecond)
	}


	fmt.Fprintf(w, "--%s--\r\n", boundary)
	flusher.Flush()
}

func main() {
	http.HandleFunc("/table-data", streamHandler)
	// send index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/index.html")
	})
	fmt.Println("Listening at http://localhost:8080")


	http.ListenAndServe(":8080", nil)
}
