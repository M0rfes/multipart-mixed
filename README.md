# Streaming JSON Data with Multipart/Mixed and Meros

You can stream JSON chunks using `Content-Type: multipart/mixed; boundary=boundary123abc` headers. This allows sending multiple JSON objects over a single HTTP connection as they become available, instead of waiting for all data to be processed.

## How Multipart Streaming Works

Multipart streaming uses HTTP chunked transfer encoding with multipart boundaries to separate individual JSON objects. Each JSON chunk is wrapped in a multipart section:

```
--boundary123abc
Content-Type: application/json

{"id": "row_1", "data": "processed", "status": "complete"}
--boundary123abc
Content-Type: application/json

{"id": "row_2", "data": "processed", "status": "complete"}
--boundary123abc--
```

You also need `Transfer-Encoding: chunked` to send the response without knowing its total size beforehand. This allows the server to start sending data immediately as it becomes available.

## Server Implementation

Here's a Go server that demonstrates streaming table data where each row requires CPU-intensive calculations:

```go
func streamTableData(w http.ResponseWriter, r *http.Request) {
    boundary := "boundary123abc"
    w.Header().Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))
    w.Header().Set("Transfer-Encoding", "chunked")
    
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
        return
    }
    
    // Process each row individually
    for _, row := range dataRows {
        // CPU-bound operation for each row
        processedData := performHeavyCalculation(row)
        
        json, err := json.Marshal(processedData)
        if err != nil {
            continue
        }
        
        // Send this row immediately
        fmt.Fprintf(w, "--%s\r\n", boundary)
        fmt.Fprint(w, "Content-Type: application/json\r\n\r\n")
        fmt.Fprint(w, string(json))
        fmt.Fprint(w, "\r\n")
        flusher.Flush()
    }
    
    // End the multipart stream
    fmt.Fprintf(w, "--%s--\r\n", boundary)
    flusher.Flush()
}
```

## Client Implementation with Meros

Meros is a JavaScript library that parses multipart streams:

```javascript
import { meros } from 'https://cdn.skypack.dev/meros';

async function streamMultipartWithMeros(url, onPart) {
    const response = await fetch(url, {
        headers: { Accept: "multipart/mixed" }
    });
    
    const parts = await meros(response);
    for await (const part of parts) {
        onPart(part.body);
    }
}

// Usage
const tableRows = [];
streamMultipartWithMeros("/table-data", (rowData) => {
    tableRows.push(rowData);
    updateTable(); // Update UI immediately
});
```

## Table Data Scenario

Consider a scenario where calculating data for each table row involves CPU-bound operations like:
- Complex mathematical computations
- Data aggregations from multiple sources  
- Heavy data processing or transformations

Without streaming, users wait for all rows to be calculated before seeing any results. With multipart streaming, each row appears as soon as its calculation completes.

## Multiple Data Sources Example

You can also stream data from different sources concurrently:

```go
func streamMultipleSources(w http.ResponseWriter, r *http.Request) {
    boundary := "boundary123abc"
    w.Header().Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))
    w.Header().Set("Transfer-Encoding", "chunked")
    
    flusher := w.(http.Flusher)
    
    postCh := make(chan string)
    commentCh := make(chan string)
    userCh := make(chan string)
    
    go fetchPosts(postCh)
    go fetchComments(commentCh)
    go fetchUsers(userCh)
    
    sendPart := func(jsonData string) {
        fmt.Fprintf(w, "--%s\r\n", boundary)
        fmt.Fprint(w, "Content-Type: application/json\r\n\r\n")
        fmt.Fprint(w, jsonData)
        fmt.Fprint(w, "\r\n")
        flusher.Flush()
    }
    
    // Send data as it becomes available from any source
    for {
        select {
        case post := <-postCh:
            if post == "" { // Channel closed
                postCh = nil
            } else {
                sendPart(post)
            }
        case comment := <-commentCh:
            if comment == "" {
                commentCh = nil
            } else {
                sendPart(comment)
            }
        case user := <-userCh:
            if user == "" {
                userCh = nil
            } else {
                sendPart(user)
            }
        }
        
        if postCh == nil && commentCh == nil && userCh == nil {
            break
        }
    }
    
    fmt.Fprintf(w, "--%s--\r\n", boundary)
    flusher.Flush()
}
```

## Client-Side State Management

When receiving data from multiple sources, use Maps to track what's loaded:

```javascript
const posts = new Map();
const comments = new Map(); 
const users = new Map();

streamMultipartWithMeros("/stream", (data) => {
    if (data.type === "post") {
        posts.set(data.id, data);
    } else if (data.type === "comment") {
        comments.set(data.id, data);
    } else if (data.type === "user") {
        users.set(data.id, data);
    }
    
    renderData(); // Update UI with current state
});

function renderData() {
    posts.forEach(post => {
        const postComments = post.comments.map(commentId => {
            const comment = comments.get(commentId);
            if (!comment) return `<div>Loading comment...</div>`;
            
            const user = users.get(comment.userId);
            const userName = user ? user.name : "Loading user...";
            
            return `<div>${comment.text} - ${userName}</div>`;
        });
        
        // Render post with available data
    });
}
```

## When to Use Multipart Streaming

Use multipart streaming when:
- Individual data pieces can be processed independently
- Computing each piece takes significant time
- Users benefit from seeing partial results immediately
- You're aggregating data from multiple slow sources

## Benefits

- **Immediate feedback**: Users see data as it becomes available
- **Better perceived performance**: Progressive loading feels faster
- **Memory efficiency**: No need to buffer large responses
- **Error resilience**: Partial failures don't block successful data

## Browser Considerations

- Browsers limit concurrent HTTP connections per domain
- Works with standard HTTP infrastructure
- Compatible with CDNs and load balancers
- Meros library handles parsing complexity

Multipart streaming provides a straightforward way to improve user experience when dealing with slow data generation or multiple data sources.

---

#multipart #streaming #HTTP #JSON #JavaScript #Go #webdevelopment #meros #chunked #API #realtime #performance #progressiveloading #webapi #nodejs #frontend #backend

https://github.com/user-attachments/assets/3775994e-e5c1-4e9b-b5e8-bab3d8168168


https://github.com/user-attachments/assets/e8f9d11a-b24f-42cc-b3cd-25a065d275d5


