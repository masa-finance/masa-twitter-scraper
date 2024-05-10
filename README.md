# Twitter Scraper

[![Go Reference](https://pkg.go.dev/badge/github.com/masa-finance/masa-twitter-scraper.svg)](https://pkg.go.dev/github.com/masa-finance/masa-twitter-scraper)

Twitter's API is annoying to work with, and has lots of limitations —
luckily their frontend (JavaScript) has it's own API, which I reverse-engineered.
No API rate limits. No tokens needed. No restrictions. Extremely fast.

You can use this library to get the text of any user's Tweets trivially.

## Installation

```shell
go get -u github.com/masa-finance/masa-twitter-scraper
```

## Usage

### Authentication

Now all methods require authentication!

#### Login

```golang
err := scraper.Login("username", "password")
```

Use username to login, not email!
But if you have email confirmation, use email address in addition:

```golang
err := scraper.Login("username", "password", "email")
```

If you have two-factor authentication, use code:

```golang
err := scraper.Login("username", "password", "code")
```

Status of login can be checked with:

```golang
scraper.IsLoggedIn()
```

Logout (clear session):

```golang
scraper.Logout()
```

If you want save session between restarts, you can save cookies with `scraper.GetCookies()` and restore with `scraper.SetCookies()`.

For example, save cookies:

```golang
cookies := scraper.GetCookies()
// serialize to JSON
js, _ := json.Marshal(cookies)
// save to file
f, _ = os.Create("cookies.json")
f.Write(js)
```

and load cookies:

```golang
f, _ := os.Open("cookies.json")
// deserialize from JSON
var cookies []*http.Cookie
json.NewDecoder(f).Decode(&cookies)
// load cookies
scraper.SetCookies(cookies)
// check login status
scraper.IsLoggedIn()
```

#### Open account

If you don't want to use your account, you can try login as a Twitter app:

```golang
err := scraper.LoginOpenAccount()
```

### Get user tweets

```golang
package main

import (
    "context"
    "fmt"
    twitterscraper "github.com/masa-finance/masa-twitter-scraper"
)

func main() {
    scraper := twitterscraper.New()
    err := scraper.LoginOpenAccount()
    if err != nil {
        panic(err)
    }
    for tweet := range scraper.GetTweets(context.Background(), "Twitter", 50) {
        if tweet.Error != nil {
            panic(tweet.Error)
        }
        fmt.Println(tweet.Text)
    }
}
```

It appears you can ask for up to 50 tweets.

### Get single tweet

```golang
package main

import (
    "fmt"

    twitterscraper "github.com/masa-finance/masa-twitter-scraper"
)

func main() {
    scraper := twitterscraper.New()
    err := scraper.Login(username, password)
    if err != nil {
        panic(err)
    }
    tweet, err := scraper.GetTweet("1328684389388185600")
    if err != nil {
        panic(err)
    }
    fmt.Println(tweet.Text)
}
```

### Search tweets by query standard operators

Now the search only works for authenticated users!

Tweets containing “twitter” and “scraper” and “data“, filtering out retweets:

```golang
package main

import (
    "context"
    "fmt"
    twitterscraper "github.com/masa-finance/masa-twitter-scraper"
)

func main() {
    scraper := twitterscraper.New()
    err := scraper.Login(username, password)
    if err != nil {
        panic(err)
    }
    for tweet := range scraper.SearchTweets(context.Background(),
        "twitter scraper data -filter:retweets", 50) {
        if tweet.Error != nil {
            panic(tweet.Error)
        }
        fmt.Println(tweet.Text)
    }
}
```

The search ends if we have 50 tweets.

See [Rules and filtering](https://developer.twitter.com/en/docs/tweets/rules-and-filtering/overview/standard-operators) for build standard queries.


#### Set search mode

```golang
scraper.SetSearchMode(twitterscraper.SearchLatest)
```

Options:

* `twitterscraper.SearchTop` - default mode
* `twitterscraper.SearchLatest` - live mode
* `twitterscraper.SearchPhotos` - image mode
* `twitterscraper.SearchVideos` - video mode
* `twitterscraper.SearchUsers` - user mode

### Get profile

```golang
package main

import (
    "fmt"
    twitterscraper "github.com/masa-finance/masa-twitter-scraper"
)

func main() {
    scraper := twitterscraper.New()
    scraper.LoginOpenAccount()
    profile, err := scraper.GetProfile("Twitter")
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", profile)
}
```

### Search profiles by query

```golang
package main

import (
    "context"
    "fmt"
    twitterscraper "github.com/masa-finance/masa-twitter-scraper"
)

func main() {
    scraper := twitterscraper.New().SetSearchMode(twitterscraper.SearchUsers)
    err := scraper.Login(username, password)
    if err != nil {
        panic(err)
    }
    for profile := range scraper.SearchProfiles(context.Background(), "Twitter", 50) {
        if profile.Error != nil {
            panic(profile.Error)
        }
        fmt.Println(profile.Name)
    }
}
```

### Get trends

```golang
package main

import (
    "fmt"
    twitterscraper "github.com/masa-finance/masa-twitter-scraper"
)

func main() {
    scraper := twitterscraper.New()
    trends, err := scraper.GetTrends()
    if err != nil {
        panic(err)
    }
    fmt.Println(trends)
}
```

### Use Proxy

Support HTTP(s) and SOCKS5 proxy

#### with HTTP

```golang
err := scraper.SetProxy("http://localhost:3128")
if err != nil {
    panic(err)
}
```

#### with SOCKS5

```golang
err := scraper.SetProxy("socks5://localhost:1080")
if err != nil {
    panic(err)
}
```

### Delay requests

Add delay between API requests (in seconds)

```golang
scraper.WithDelay(5)
```

### Load timeline with tweet replies

```golang
scraper.WithReplies(true)
```
