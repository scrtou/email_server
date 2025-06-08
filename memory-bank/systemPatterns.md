# System Patterns *Optional*

This file documents recurring patterns and standards used in the project.
It is optional, but recommended to be updated as the project evolves.
2025-06-08 11:10:04 - Log of updates made.

*

## Coding Patterns

*   

## Architectural Patterns

*   

## Testing Patterns

*
---
### Data Models
[2025-06-08 11:10:31] - 为邮件收件箱功能定义了核心数据模型。这些模型将在 Go 后端的 `models` 包中实现。

**`Email` Model:**
```go
type Email struct {
    ID          uint      `json:"id"`
    MessageID   string    `json:"messageId"`
    Subject     string    `json:"subject"`
    From        []EmailAddress `json:"from"`
    To          []EmailAddress `json:"to"`
    Cc          []EmailAddress `json:"cc"`
    Date        time.Time `json:"date"`
    Snippet     string    `json:"snippet"`
    Body        string    `json:"body"`
    HTMLBody    string    `json:"htmlBody"`
    IsRead      bool      `json:"isRead"`
    HasAttachment bool    `json:"hasAttachment"`
    Attachments []Attachment `json:"attachments"`
}
```

**`EmailAddress` Model:**
```go
type EmailAddress struct {
    Name    string `json:"name"`
    Address string `json:"address"`
}
```

**`Attachment` Model:**
```go
type Attachment struct {
    Filename string `json:"filename"`
    MimeType string `json:"mimeType"`
    Size     int64  `json:"size"`
    // Content is handled separately to avoid loading large files into memory
    ContentID string `json:"contentId"` // Used for inline images
}
```
---
### System Pattern: API Pagination with Frontend Infinite Scrolling

**Context:**
When displaying a large, potentially unbounded list of items (e.g., emails, posts, notifications), loading all data at once is inefficient and leads to poor performance and user experience.

**Pattern:**
This pattern combines backend API pagination with a frontend infinite scrolling mechanism.

1.  **Backend (API):**
    *   The API endpoint responsible for fetching the list (e.g., `GET /api/v1/inbox`) accepts pagination parameters, typically `page` (the page number to retrieve) and `pageSize` (the number of items per page).
    *   The API response includes the list of items for the requested page (`data` or a named array like `emails`) and the `total` number of items available on the server.
    *   **Example Response:**
        ```json
        {
          "emails": [ ... list of email objects ... ],
          "total": 1250
        }
        ```

2.  **Frontend (UI/State Management):**
    *   **State:** A client-side state store (e.g., Pinia, Redux) manages the list of items, the current `page`, the `pageSize`, the `total` count from the server, and a boolean flag like `hasMore` to indicate if more items can be loaded.
    *   **Initial Load:** On component mount, the frontend fetches the first page (`page: 1`).
    *   **Infinite Scroll:** A scroll event listener is attached to the list container or the window. When the user scrolls near the bottom of the list, the frontend checks if `isLoading` is false and `hasMore` is true.
    *   **Loading More:** If conditions are met, the `page` number is incremented, and another API request is made to fetch the next page. The new items are appended to the existing list in the state store.
    *   **End of List:** The `hasMore` flag is set to `false` when the number of items in the local list equals or exceeds the `total` count from the server. This prevents further API calls.
    *   **UI Feedback:** Loading indicators are shown during the initial fetch and subsequent "load more" operations. A message is displayed at the end of the list when `hasMore` is false.

**Benefits:**
*   **Performance:** Drastically improves initial page load time.
*   **Scalability:** Handles very large datasets gracefully.
*   **User Experience:** Provides a smooth, seamless browsing experience without manual page clicks.

**Implementation Files:**
*   **Backend Handler:** [`src/backend/handlers/email.go`](src/backend/handlers/email.go)
*   **Backend IMAP Client:** [`src/backend/integrations/imap_client.go`](src/backend/integrations/imap_client.go)
*   **Frontend View:** [`src/frontend/src/views/InboxView.vue`](src/frontend/src/views/InboxView.vue)
*   **Frontend Store:** [`src/frontend/src/stores/inbox.js`](src/frontend/src/stores/inbox.js)
---
### System Pattern: OAuth2 Authorization Code Flow for Third-Party Email Integration

**Context:**
To allow users to connect their external email accounts (e.g., from Google, Microsoft) without storing their passwords, the system must implement the OAuth2 Authorization Code grant type. This is a secure, redirect-based flow that delegates authentication to the third-party provider.

**Pattern:**
The flow involves a sequence of interactions between the user's browser, our backend, and the third-party OAuth2 provider.

1.  **Initiation (Frontend/Backend):**
    *   The user clicks a "Connect with [Provider]" button on the frontend.
    *   The frontend calls a backend endpoint (e.g., `GET /api/v1/oauth2/connect/{provider}`).
    *   The backend, using the `golang.org/x/oauth2` library, constructs the unique authorization URL for the provider, including parameters like `client_id`, `redirect_uri`, `response_type=code`, `scope`, and a `state` parameter for CSRF protection.
    *   The backend redirects the user's browser to this authorization URL.

2.  **User Authorization (Third-Party Provider):**
    *   The user is presented with the provider's consent screen, asking for permission to grant the requested scopes (e.g., read emails).
    *   Upon approval, the provider redirects the user's browser back to the `redirect_uri` specified in step 1. The redirect includes an `authorization_code` and the original `state` parameter in the query string.

3.  **Callback and Token Exchange (Backend):**
    *   The redirect URI points to a backend callback handler (e.g., `GET /api/v1/oauth2/callback/{provider}`).
    *   The handler first validates that the received `state` parameter matches the one generated in step 1 to prevent CSRF attacks.
    *   The handler then exchanges the `authorization_code` for an `access_token` and a `refresh_token` by making a server-to-server POST request to the provider's token endpoint. The `golang.org/x/oauth2` library handles this exchange.
    *   The received tokens (access and refresh) are encrypted and stored securely in the database, associated with the user's account and the specific email account they are connecting.

4.  **Completion (Backend/Frontend):**
    *   After successfully storing the tokens, the backend redirects the user back to a specific page in the frontend application (e.g., `/settings/email-accounts`).
    *   The frontend can then display a success message, indicating that the account has been connected. The application can now use the stored access token to make API calls to the provider on the user's behalf.

**Implementation Files:**
*   **Backend Handler:** [`src/backend/handlers/oauth2.go`](src/backend/handlers/oauth2.go) (to be created)
*   **Database Models:** [`src/backend/models/user.go`](src/backend/models/user.go), [`src/backend/models/email_account.go`](src/backend/models/email_account.go) (to be updated)
*   **Frontend View:** [`src/frontend/src/views/OAuth2Callback.vue`](src/frontend/src/views/OAuth2Callback.vue) (for handling the final redirect)