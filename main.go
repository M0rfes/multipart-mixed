package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Post struct {
	ID       string   `json:"id"`
	Data     string   `json:"data"`
	Comments []string `json:"comments"`
}

type Comment struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	User string `json:"user"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	// Mock data
	posts = []Post{
		{ID: "p1", Data: "Hello", Comments: []string{"c1", "c2"}},
		{ID: "p2", Data: "World", Comments: []string{"c3", "c4"}},
		{ID: "p3", Data: "Hello", Comments: []string{"c5", "c6"}},
		{ID: "p4", Data: "World", Comments: []string{"c7", "c8"}},
		{ID: "p5", Data: "Hello", Comments: []string{"c9", "c10"}},
		{ID: "p6", Data: "World", Comments: []string{"c11", "c12"}},
		{ID: "p7", Data: "Hello", Comments: []string{"c13", "c14"}},
		{ID: "p8", Data: "World", Comments: []string{"c15", "c16"}},
		{ID: "p9", Data: "Hello", Comments: []string{"c17", "c18"}},
		{ID: "p10", Data: "World", Comments: []string{"c19", "c20"}},
	}

	comments = []Comment{
		{ID: "c1", Text: "Great!", User: "u1"},
		{ID: "c2", Text: "Good job!", User: "u2"},
		{ID: "c3", Text: "Thanks!", User: "u3"},
		{ID: "c4", Text: "Awesome!", User: "u4"},
		{ID: "c5", Text: "Thanks!", User: "u5"},
		{ID: "c6", Text: "Thanks!", User: "u6"},
		{ID: "c7", Text: "Thanks!", User: "u7"},
		{ID: "c8", Text: "Thanks!", User: "u8"},
		{ID: "c9", Text: "Thanks!", User: "u9"},
		{ID: "c10", Text: "Thanks!", User: "u10"},
		{ID: "c11", Text: "Thanks!", User: "u11"},
		{ID: "c12", Text: "Thanks!", User: "u12"},
		{ID: "c13", Text: "Thanks!", User: "u13"},
		{ID: "c14", Text: "Thanks!", User: "u14"},
		{ID: "c15", Text: "Thanks!", User: "u15"},
		{ID: "c16", Text: "Thanks!", User: "u16"},
		{ID: "c17", Text: "Thanks!", User: "u17"},
		{ID: "c18", Text: "Thanks!", User: "u18"},
		{ID: "c19", Text: "Thanks!", User: "u19"},
		{ID: "c20", Text: "Thanks!", User: "u20"},
	}

	users = []User{
		{ID: "u1", Name: "Alice"},
		{ID: "u2", Name: "Bob"},
		{ID: "u3", Name: "Charlie"},
		{ID: "u4", Name: "David"},
		{ID: "u5", Name: "Eve"},
		{ID: "u6", Name: "Frank"},
		{ID: "u7", Name: "George"},
		{ID: "u8", Name: "Hannah"},
		{ID: "u9", Name: "Isaac"},
		{ID: "u10", Name: "James"},
		{ID: "u11", Name: "John"},
		{ID: "u12", Name: "Kate"},
		{ID: "u13", Name: "Liam"},
		{ID: "u14", Name: "Mia"},
		{ID: "u15", Name: "Noah"},
		{ID: "u16", Name: "Olivia"},
		{ID: "u17", Name: "Patrick"},
		{ID: "u18", Name: "Quinn"},
		{ID: "u19", Name: "Ryan"},
		{ID: "u20", Name: "Sarah"},
	}
)

func getPosts(ch chan<- string) {
	for _, post := range posts {
		postMap := make(map[string]any)
		postMap["type"] = "post"
		postMap[post.ID] = post
		postJSON, err := json.Marshal(postMap)
		if err != nil {
			fmt.Println("Error marshalling post:", err)
			continue
		}
		time.Sleep(500 * time.Millisecond)
		ch <- string(postJSON)
	}
	close(ch)
}

func getComments(ch chan<- string) {
	for i := 0; i < len(comments); i += 2 {
		commentMap := make(map[string]any)
		commentMap["type"] = "comment"
		commentMap[comments[i].ID] = comments[i]
		commentMap[comments[i+1].ID] = comments[i+1]
		commentJSON, err := json.Marshal(commentMap)
		if err != nil {
			fmt.Println("Error marshalling comments:", err)
			continue
		}
		time.Sleep(600 * time.Millisecond)
		ch <- string(commentJSON)
	}
	close(ch)
}

func getUsers(ch chan<- string) {
	for i := 0; i < len(users); i += 1 {
		userMap := make(map[string]any)
		userMap["type"] = "user"
		userMap[users[i].ID] = users[i]
		userJSON, err := json.Marshal(userMap)
		if err != nil {
			fmt.Println("Error marshalling users:", err)
			continue
		}
		time.Sleep(700 * time.Millisecond)
		ch <- string(userJSON)
	}
	close(ch)
}

func streamHandler(w http.ResponseWriter, r *http.Request) {

	boundary := "boundary123abc"
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(200)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	var mu sync.Mutex
	sendPart := func(jsonPayload string) {
		defer mu.Unlock()
		fmt.Fprintf(w, "--%s\r\n", boundary)
		fmt.Fprint(w, "Content-Type: application/json\r\n\r\n")
		fmt.Fprint(w, jsonPayload)
		fmt.Fprint(w, "\r\n")
		flusher.Flush()
		// need this to make sure the client is ready to receive the next part
		// TODO: find a better way to do this
		time.Sleep(10 * time.Millisecond)
	}

	postCh := make(chan string)
	commentCh := make(chan string)
	userCh := make(chan string)
	go getPosts(postCh)
	go getComments(commentCh)
	go getUsers(userCh)

	doneCh := make(chan bool)

	go func() {
		postClosed := false
		commentClosed := false
		userClosed := false

		for !postClosed || !commentClosed || !userClosed {
			select {
			case post, ok := <-postCh:
				if !ok {
					postClosed = true
				} else {
					mu.Lock()
					sendPart(post)
				}
			case comment, ok := <-commentCh:
				if !ok {
					commentClosed = true
				} else {
					mu.Lock()
					sendPart(comment)
				}
			case user, ok := <-userCh:
				if !ok {
					userClosed = true
				} else {
					mu.Lock()
					sendPart(user)
				}
			}
		}
		doneCh <- true
	}()

	<-doneCh
	fmt.Fprintf(w, "--%s--\r\n", boundary)
	flusher.Flush()
}

func main() {
	http.HandleFunc("/stream", streamHandler)
	// send index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/index.html")
	})
	fmt.Println("Listening at http://localhost:8080")


	http.ListenAndServe(":8080", nil)
}
